package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/hotnops/apeman/src/api/src/queries"
)

func (s *Server) GetAWSGroup(c *gin.Context) {
	propertyName := "groupid"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSGroup)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSGroupMembers(c *gin.Context) {
	propertyName := "groupid"
	id := c.Param(propertyName)

	node, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSGroup)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	members, err := queries.GetAWSGroupMembers(s.ctx, s.db, node)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, members)
}

func (s *Server) GetAWSGroupPolicies(c *gin.Context) {
	propertyName := "groupid"
	id := c.Param(propertyName)

	nodes, err := queries.GetPoliciesOfEntity(s.ctx, s.db, propertyName, id, aws.AWSManagedPolicy)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	var returnValue NodeResponse
	returnValue.Version = API_VERSION
	returnValue.Count = len(nodes)
	returnValue.Nodes = nodes.Slice()

	c.IndentedJSON(http.StatusOK, returnValue)
}

func (s *Server) addGroupsEndpoints(router *gin.RouterGroup) {
	router.GET("", s.GetAWSGroup)
	router.GET("members", s.GetAWSGroupMembers)
	router.GET("policies", s.GetAWSGroupPolicies)
}
