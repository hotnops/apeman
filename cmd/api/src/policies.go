package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/hotnops/apeman/src/api/src/queries"
)

func (s *Server) GenerateInlinePolicy(c *gin.Context) {
	policyHash := c.Param("policyhash")

	policyObject, err := queries.GenerateInlinePolicyObject(s.ctx, s.db, policyHash)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, policyObject)
}

func (s *Server) GenerateManagedPolicy(c *gin.Context) {
	policyId := c.Param("policyid")

	policyObject, err := queries.GenerateManagedPolicyObject(s.ctx, s.db, policyId)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, policyObject)
}

func (s *Server) GetAWSInlinePolicyNodes(c *gin.Context) {
	propertyName := "policyhash"
	policyId := c.Param(propertyName)

	policyNode, err := queries.GetAWSNodeByKindID(s.ctx, s.db, "hash", policyId, aws.AWSInlinePolicy)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	paths, err := queries.GetNodesOfPolicy(s.ctx, s.db, policyNode.ID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, paths)
}

func (s *Server) GetAWSManagedPolicyNodes(c *gin.Context) {
	propertyName := "policyid"
	policyId := c.Param(propertyName)

	policyNode, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, policyId, aws.AWSManagedPolicy)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	paths, err := queries.GetNodesOfPolicy(s.ctx, s.db, policyNode.ID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, paths)
}

func (s *Server) GetAWSPolicy(c *gin.Context) {
	propertyName := "policyid"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSManagedPolicy)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSPolicyPrincipals(c *gin.Context) {
	propertyName := "policyid"
	policyId := c.Param(propertyName)

	principals, err := queries.GetPrincipalsOfPolicy(s.ctx, s.db, policyId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, principals.Slice())

}

func (s *Server) addInlinePoliciesEndpoints(router *gin.RouterGroup) {
	router.GET("generatepolicy", s.GenerateInlinePolicy)
	router.GET("nodes", s.GetAWSInlinePolicyNodes)
}

func (s *Server) addManagedPoliciesEndpoints(router *gin.RouterGroup) {
	router.GET("", s.GetAWSPolicy)
	router.GET("generatepolicy", s.GenerateManagedPolicy)
	router.GET("principals", s.GetAWSPolicyPrincipals)
	router.GET("nodes", s.GetAWSManagedPolicyNodes)
}
