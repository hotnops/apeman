package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/analyze"
	"github.com/hotnops/apeman/src/api/src/queries"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func (s *Server) GetAWSNodes(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	var nodes []*graph.Node
	var err error
	nodes, err = queries.GetAllAWSNodes(s.ctx, s.db, queryParams)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSNodeByGraphID(c *gin.Context) {
	propertyName := "nodeid"
	idString := c.Param(propertyName)

	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	node, err := analyze.GetAWSNodeByGraphID(s.ctx, s.db, graph.ID(id))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, node)

}

func (s *Server) GetAWSNodeEdges(c *gin.Context, direction graph.Direction) {
	propertyName := "nodeid"
	idString := c.Param(propertyName)
	queryParams := c.Request.URL.Query()

	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	relationships, err := queries.GetAWSNodeEdges(s.ctx, s.db, graph.ID(id), direction, queryParams)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	var returnValue RelationshipResponse
	returnValue.Version = API_VERSION
	returnValue.Count = len(relationships)
	returnValue.Relationships = relationships

	c.IndentedJSON(http.StatusOK, returnValue)
}

func (s *Server) GetAWSNodeTags(c *gin.Context) {
	nodeIdStr := c.Param("nodeid")
	nodeId, err := strconv.ParseUint(nodeIdStr, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	tags, err := queries.GetAWSNodeTags(s.ctx, s.db, graph.ID(nodeId))

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, tags.Slice())
}

func (s *Server) GetNodeIdentityPath(c *gin.Context) {
	sourceNodeIdStr := c.Param("nodeid")
	destNodeIdStr := c.Param("destnodeid")

	sourceNodeId, err := strconv.ParseUint(sourceNodeIdStr, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	destNodeId, err := strconv.ParseInt(destNodeIdStr, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	query := fmt.Sprintf("MATCH p=shortestPath((a) - [:IdentityTransform*] -> (b)) WHERE ID(a) = %d AND ID(b) = %d RETURN p", sourceNodeId, destNodeId)
	paths, err := queries.CypherQueryPaths(s.ctx, s.db, query)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, paths)
}

func (s *Server) GetNodeShortestPath(c *gin.Context) {
	sourceNodeIdStr := c.Param("nodeid")
	destNodeIdStr := c.Param("destnodeid")

	sourceNodeId, err := strconv.ParseUint(sourceNodeIdStr, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	destNodeId, err := strconv.ParseUint(destNodeIdStr, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	paths, err := queries.CypherQueryPaths(s.ctx, s.db, fmt.Sprintf("MATCH p=(a) - [*1..3] - (b) WHERE ID(a) = %d AND ID(b) = %d RETURN p", sourceNodeId, destNodeId))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, paths)

}

func (s *Server) GetAWSNodeInboundEdges(c *gin.Context) {
	s.GetAWSNodeEdges(c, graph.DirectionInbound)
}

func (s *Server) GetAWSNodeOutboundEdges(c *gin.Context) {
	s.GetAWSNodeEdges(c, graph.DirectionOutbound)
}

func (s *Server) addNodeEndpoints(router *gin.RouterGroup) {

	router.GET("", s.GetAWSNodeByGraphID)
	router.GET("shortestpath/:destnodeid", s.GetNodeShortestPath)
	router.GET("identitypath/:destnodeid", s.GetNodeIdentityPath)
	router.GET("inboundedges", s.GetAWSNodeInboundEdges)
	router.GET("outboundedges", s.GetAWSNodeOutboundEdges)
	router.GET("tags", s.GetAWSNodeTags)
}
