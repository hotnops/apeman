package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/analyze"
	"github.com/hotnops/apeman/go/internal/queries"
	"github.com/hotnops/apeman/graphschema/aws"
)

func (s *Server) GetAWSUser(c *gin.Context) {
	propertyName := "userid"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSUser)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSUserInlinePolicy(c *gin.Context) {
	s.GetPrincipalInlinePolicy(c, "userid", c.Param("userid"))
}

func (s *Server) GetAWSUserManagedPolicies(c *gin.Context) {
	propertyName := "userid"
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

func (s *Server) GetAWSUserOutboundRoles(c *gin.Context) {

	roleId := c.Param("userid")

	//paths, err := queries.GetAWSRoleInboundRoleAssumptionPaths(s.ctx, s.db, roleId)
	query := "MATCH p=(a:AWSUser) - [:IdentityTransform* {name: 'sts:assumerole'}] -> (b:AWSRole) WHERE a.userid = '%s' AND ALL(n IN nodes(p) WHERE SINGLE(x IN nodes(p) WHERE x = n)) RETURN p"
	query = fmt.Sprintf(query, roleId)
	paths, err := queries.CypherQueryPaths(s.ctx, s.db, query)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, paths)

}

func (s *Server) GetAWSUserRSOP(c *gin.Context) {
	userId := c.Param("userid")
	node, err := queries.GetAWSNodeByKindID(s.ctx, s.db, "userid", userId, aws.AWSUser)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	paths, err := queries.GetUnresolvedOutputPaths(s.ctx, s.db, node)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	resolvedPaths, err := analyze.ResolveResourceAgainstIdentityPolicies(&analyze.ActionPathSet{}, &paths)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)

	}

	actionToPrin := analyze.ResourcePathSetToMap(*resolvedPaths)
	c.IndentedJSON(http.StatusOK, actionToPrin)
}

func (s *Server) GetAWSUserRSOPActions(c *gin.Context) {
	userId := c.Param("userid")
	node, err := queries.GetAWSNodeByKindID(s.ctx, s.db, "userid", userId, aws.AWSUser)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	paths, err := queries.GetUnresolvedOutputPaths(s.ctx, s.db, node)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	resolvedPaths, err := analyze.ResolveResourceAgainstIdentityPolicies(&analyze.ActionPathSet{}, &paths)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)

	}

	principalMap, err := analyze.GetActionMapFromPathSet(*resolvedPaths)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, principalMap)
}

func (s *Server) addUserEndpoints(user *gin.RouterGroup) {
	user.GET("", s.GetAWSUser)
	user.GET("managedpolicies", s.GetAWSUserManagedPolicies)
	user.GET("inlinepolicy", s.GetAWSUserInlinePolicy)
	user.GET("rsop", s.GetAWSUserRSOP)
	user.GET("rsop/actions", s.GetAWSUserRSOPActions)
	user.GET("outboundroles", s.GetAWSUserOutboundRoles)
}
