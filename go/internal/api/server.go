package api

// The server file is the API interface, and all parameters
// should be as generic as possible. This layer should commonly
// convert string parametes to their corresponding graph IDs
// when possible.

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/go/internal/config"
	"github.com/hotnops/apeman/go/internal/queries"
	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

var API_VERSION string = "v1.0"

type Server struct {
	db     graph.Database
	ctx    context.Context
	config dawgs.Config
}

type RelationshipResponse struct {
	Version       string                `json:"version"`
	Count         int                   `json:"count"`
	Relationships []*graph.Relationship `json:"relationships"`
}

type NodeResponse struct {
	Version string        `json:"version"`
	Count   int           `json:"count"`
	Nodes   []*graph.Node `json:"nodes"`
}

func (s *Server) GetAWSRelationshipByGraphID(c *gin.Context) {
	propertyName := "relationshipid"
	idString := c.Param(propertyName)

	id, err := strconv.Atoi(idString)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	relationship, err := queries.GetAWSRelationshipByGraphID(s.ctx, s.db, graph.ID(id))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, relationship)

}

func (s *Server) GetPrincipalInlinePolicy(c *gin.Context, propertyName string, id string) {

	nodes, err := queries.GetPoliciesOfEntity(s.ctx, s.db, propertyName, id, aws.AWSInlinePolicy)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	if nodes.Len() == 0 {
		c.IndentedJSON(http.StatusOK, nil)
	}

	c.IndentedJSON(http.StatusOK, nodes.Slice()[0])
}

func DecodeArn(encodedArn string) (string, error) {
	arn, err := base64.URLEncoding.DecodeString(encodedArn)
	if err != nil {
		return "", err
	}

	arnString := string(arn[:])
	return arnString, nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (s *Server) PostQuery(c *gin.Context) {
	query := c.PostForm("query")

	if response, err := queries.CypherQueryPaths(s.ctx, s.db, query); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

func (s *Server) Search(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	searchString := queryParams.Get("searchQuery")

	if len(searchString) < 3 {
		c.Abort()
	}

	if response, err := queries.Search(s.ctx, s.db, searchString); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

func (s *Server) AnalyzeIdentityTransforms(c *gin.Context) {
	wg := sync.WaitGroup{}

	// go func1() {
	// 	defer wg.Done()
	// 	queries.CreateAssumeRoleEdges(s.ctx, s.db, &wg)
	// }()
	// go
	wg.Add(3)
	go func() {
		defer wg.Done()
		queries.CreateAssumeRoleEdges(s.ctx, s.db)
	}()
	go func() {
		defer wg.Done()
		queries.CreateUpdateAssumeRoleEdges(s.ctx, s.db)
	}()
	go func() {
		defer wg.Done()
		queries.CreateCreateAccessKeyEdges(s.ctx, s.db)
	}()

	wg.Wait()
	c.Done()
}

func (s *Server) GetNodePermissionPath(c *gin.Context) {
	sourceNodeIdStr := c.Param("sourcenodeid")
	destNodeIdStr := c.Param("destnodeid")

	queryParams := c.Request.URL.Query()
	action := queryParams.Get("action")

	sourceNodeId, err := strconv.ParseUint(sourceNodeIdStr, 10, 32)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	destNodeId, err := strconv.ParseInt(destNodeIdStr, 10, 32)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	paths, err := queries.GetNodePermissionPath(s.ctx, s.db, graph.ID(sourceNodeId), graph.ID(destNodeId), action)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, paths)
}

func (s *Server) handleRequests() {
	router := gin.Default()
	router.Use(corsMiddleware())
	s.addRoleEndpoints(router.Group("/roles/:roleid"))
	s.addUserEndpoints(router.Group("/users/:userid"))
	s.addResourceEndpoints(router.Group("/resources/:arn"))
	s.addStatementEndpoints(router.Group("/statements/:statementhash"))
	s.addManagedPoliciesEndpoints(router.Group("/managedpolicies/:policyid"))
	s.addInlinePoliciesEndpoints(router.Group("/inlinepolicies/:policyhash"))
	s.addNodeEndpoints(router.Group("/nodes/:nodeid"))
	s.addGroupsEndpoints(router.Group("/groups/:groupid"))
	s.addAccountsEndpoints(router.Group("/accounts/:accountid"))
	s.addActionsEndpoints(router.Group("/actions/:actionname"))

	router.GET("/nodes", s.GetAWSNodes)
	router.GET("/accounts", s.GetAWSAccountIDs)

	router.GET("/permissionpath/:sourcenodeid/:destnodeid", s.GetNodePermissionPath)
	router.GET("/relationship/:relationshipid", s.GetAWSRelationshipByGraphID)
	router.GET("/analyze/identitytransforms", s.AnalyzeIdentityTransforms)
	router.GET("/search", s.Search)
	router.POST("/query", s.PostQuery)
	router.Run("0.0.0.0:4400")
}

func (s *Server) InitializeServer() {
	var err error
	s.ctx = context.Background()

	bhCfg := config.Configuration{
		Version:     1,
		BindAddress: "0.0.0.0:8080",
		MetricsPort: ":2112",
		RootURL:     "http://127.0.0.1:8080",
		WorkDir:     "/opt/apeman/work",
		LogLevel:    "INFO",
		LogPath:     "apeman.log",
		TLS: config.TLSConfiguration{
			CertFile: "",
			KeyFile:  "",
		},
		Database: config.DatabaseConfiguration{
			Connection: "user=apeman password=apeman dbname=apeman host=app-db",
		},
		Neo4J: config.DatabaseConfiguration{
			Connection: "neo4j://a:b@neo4j:7687/",
		},
	}

	s.config = dawgs.Config{
		DriverCfg:            bhCfg.Neo4J.Neo4jConnectionString(),
		TraversalMemoryLimit: size.Size(bhCfg.TraversalMemoryLimit) * size.Gibibyte,
	}

	s.db, err = dawgs.Open(neo4j.DriverName, s.config)
	if err != nil {
		log.Fatalf("Failed to open graph database")
	}
}

func (s *Server) Start() {
	s.handleRequests()
}
