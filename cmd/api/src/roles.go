package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/analyze"
	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/hotnops/apeman/src/api/src/queries"
)

func (s *Server) GenerateAssumeRolePolicy(c *gin.Context) {
	propertyName := "roleid"
	id := c.Param(propertyName)

	policy, err := queries.GenerateAssumeRolePolicy(s.ctx, s.db, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, policy)

}

func (s *Server) GetAWSRole(c *gin.Context) {
	propertyName := "roleid"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSRole)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSRoleInlinePolicy(c *gin.Context) {
	s.GetPrincipalInlinePolicy(c, "roleid", c.Param("roleid"))
}

func (s *Server) GetAWSRoleManagedPolicies(c *gin.Context) {
	propertyName := "roleid"
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

func (s *Server) GetAWSRoleOutboundRoles(c *gin.Context) {

	roleId := c.Param("roleid")

	//paths, err := queries.GetAWSRoleInboundRoleAssumptionPaths(s.ctx, s.db, roleId)
	query := "MATCH p=(a:AWSRole) - [:IdentityTransform*] -> (b:AWSRole) WHERE a.roleid = '%s' AND ALL(n IN nodes(p) WHERE SINGLE(x IN nodes(p) WHERE x = n)) RETURN p"
	query = fmt.Sprintf(query, roleId)
	paths, err := queries.CypherQueryPaths(s.ctx, s.db, query)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, paths)

}

func (s *Server) GetInboundRoles(c *gin.Context) {
	roleId := c.Param("roleid")

	//paths, err := queries.GetAWSRoleInboundRoleAssumptionPaths(s.ctx, s.db, roleId)
	paths, err := queries.GetInboundRolePaths(s.ctx, s.db, roleId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, paths)
}

func (s *Server) GetAWSRoleRSOP(c *gin.Context) {
	roleId := c.Param("roleid")
	node, err := queries.GetAWSNodeByKindID(s.ctx, s.db, "roleid", roleId, aws.AWSRole)
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

func (s *Server) GetAWSRoleRSOPActions(c *gin.Context) {
	roleId := c.Param("roleid")
	node, err := queries.GetAWSNodeByKindID(s.ctx, s.db, "roleid", roleId, aws.AWSRole)
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

func (s *Server) GetAWSRoleRSOPPrincipals(c *gin.Context) {
	roleId := c.Param("roleid")
	node, err := queries.GetAWSNodeByKindID(s.ctx, s.db, "roleid", roleId, aws.AWSRole)
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

	principalMap := analyze.GetResourceArnsFromActionSet(*resolvedPaths)
	c.IndentedJSON(http.StatusOK, principalMap)
}

func (s *Server) addRoleEndpoints(roles *gin.RouterGroup) {
	roles.GET("", s.GetAWSRole)
	roles.GET("managedpolicies", s.GetAWSRoleManagedPolicies)
	roles.GET("inlinepolicy", s.GetAWSRoleInlinePolicy)
	roles.GET("generateassumerolepolicy", s.GenerateAssumeRolePolicy)
	roles.GET("inboundroles", s.GetInboundRoles)
	roles.GET("outboundroles", s.GetAWSRoleOutboundRoles)
	roles.GET("rsop", s.GetAWSRoleRSOP)
	roles.GET("rsop/principals", s.GetAWSRoleRSOPPrincipals)
	roles.GET("rsop/actions", s.GetAWSRoleRSOPActions)
}
