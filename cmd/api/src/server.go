package api

// The server file is the API interface, and all parameters
// should be as generic as possible. This layer should commonly
// convert string parametes to their corresponding graph IDs
// when possible.

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/analyze"
	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/hotnops/apeman/src/api/src/queries"
	"github.com/hotnops/apeman/src/config"
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

	id, err := strconv.Atoi(idString)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	node, err := analyze.GetAWSNodeByGraphID(s.ctx, s.db, graph.ID(id))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, node)

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

func (s *Server) GetAWSNodeInboundEdges(c *gin.Context) {
	s.GetAWSNodeEdges(c, graph.DirectionInbound)
}

func (s *Server) GetAWSNodeOutboundEdges(c *gin.Context) {
	s.GetAWSNodeEdges(c, graph.DirectionOutbound)
}

func (s *Server) GetAWSNodeEdges(c *gin.Context, direction graph.Direction) {
	propertyName := "nodeid"
	idString := c.Param(propertyName)
	queryParams := c.Request.URL.Query()

	id, err := strconv.Atoi(idString)
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

func (s *Server) GetAWSRole(c *gin.Context) {
	propertyName := "roleid"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSRole)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSUser(c *gin.Context) {
	propertyName := "userid"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSUser)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetUserRSOP(c *gin.Context) {
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

func (s *Server) GetAWSAccount(c *gin.Context) {
	propertyName := "account_id"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSAccount)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSAccountServices(c *gin.Context) {
	propertyName := "account_id"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSAccountServices(s.ctx, s.db, id)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) GetAWSAccountIDs(c *gin.Context) {
	nodes, err := queries.GetAWSAccountIDs(s.ctx, s.db)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
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

func (s *Server) GetRoleManagedPolicies(c *gin.Context) {
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

func (s *Server) GetRoleInlinePolicies(c *gin.Context) {
	propertyName := "roleid"
	id := c.Param(propertyName)

	nodes, err := queries.GetPoliciesOfEntity(s.ctx, s.db, propertyName, id, aws.AWSInlinePolicy)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	var returnValue NodeResponse
	returnValue.Version = API_VERSION
	returnValue.Count = len(nodes)
	returnValue.Nodes = nodes.Slice()

	c.IndentedJSON(http.StatusOK, returnValue)
}

func (s *Server) GetUserInlinePolicies(c *gin.Context) {
	propertyName := "userid"
	id := c.Param(propertyName)

	nodes, err := queries.GetPoliciesOfEntity(s.ctx, s.db, propertyName, id, aws.AWSInlinePolicy)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	var returnValue NodeResponse
	returnValue.Version = API_VERSION
	returnValue.Count = len(nodes)
	returnValue.Nodes = nodes.Slice()

	c.IndentedJSON(http.StatusOK, returnValue)
}

func (s *Server) GetUserManagedPolicies(c *gin.Context) {
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

func (s *Server) GetGroupPolicies(c *gin.Context) {
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

func (s *Server) GetAWSResourceInboundPermissions(c *gin.Context) {
	propertyName := "arn"
	encodedArn := c.Param(propertyName)
	action := c.Query("actionName")
	if action != "" {
		s.GetAWSInboundPrincipalsWithActionOnArn(c, action)
	} else {
		if arnString, err := DecodeArn((encodedArn)); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		} else {
			// All the paths that act on this resource
			identityPaths, err := queries.GetAllUnresolvedIdentityPolicyPathsOnArn(s.ctx, s.db, arnString)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
			}
			// Filter through
			resolvedPaths, err := analyze.ResolveResourceAgainstIdentityPolicies(&analyze.ActionPathSet{}, identityPaths)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
			}
			if resolvedPaths == nil {
				log.Print("No paths found")
				c.IndentedJSON(http.StatusOK, nil)
			} else {
				prinToActionMap := analyze.ActionPathSetToMap(*resolvedPaths)
				if err != nil {
					c.AbortWithError(http.StatusBadRequest, err)
				}
				c.IndentedJSON(http.StatusOK, prinToActionMap)
			}
		}
	}

}

func (s *Server) GetAWSResourceInboundPermissionsPrincipals(c *gin.Context) {
	propertyName := "arn"
	encodedArn := c.Param(propertyName)
	principalMap := map[string]bool{}
	principals := []string{}

	if arnString, err := DecodeArn((encodedArn)); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	} else {
		// All the paths that act on this resource
		identityPaths, err := queries.GetAllUnresolvedIdentityPolicyPathsOnArn(s.ctx, s.db, arnString)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		// Filter through
		resolvedPaths, err := analyze.ResolveResourceAgainstIdentityPolicies(&analyze.ActionPathSet{}, identityPaths)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		if resolvedPaths == nil {
			log.Print("No paths found")
			c.IndentedJSON(http.StatusOK, nil)
		} else {
			for _, actionPath := range *resolvedPaths {
				// Add only if not already in list
				principalMap[actionPath.PrincipalArn] = true
			}

			for key := range principalMap {
				principals = append(principals, key)
			}

			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
			}
			// Return all the keys of the principals
			c.IndentedJSON(http.StatusOK, principals)
		}
	}
}

func (s *Server) GetActionsWithPrincipalOnResource(c *gin.Context) {
	encodedResourceArn := c.Param("arn")
	encodedPrincipalArn := c.Param("principalArn")

	resourceArn, err := DecodeArn(encodedResourceArn)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	principalArn, err := DecodeArn(encodedPrincipalArn)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	identityPaths, err := queries.GetAllUnresolvedIdentityPolicyPathsOnArnFromArn(s.ctx, s.db, resourceArn, principalArn)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	// Filter through
	resolvedPaths, err := analyze.ResolveResourceAgainstIdentityPolicies(&analyze.ActionPathSet{}, identityPaths)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	if resolvedPaths == nil {
		log.Print("No paths found")
		c.IndentedJSON(http.StatusOK, nil)
	} else {
		actions := []string{}

		for _, actionPath := range *resolvedPaths {
			actions = append(actions, actionPath.Action)
		}
		c.IndentedJSON(http.StatusOK, actions)
	}
}

func (s *Server) GetAWSInboundPrincipalsWithActionOnArn(c *gin.Context, actionName string) {
	propertyName := "arn"
	encodedArn := c.Param(propertyName)

	if arnString, err := DecodeArn((encodedArn)); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	} else {
		// All the paths that act on this resource
		identityPaths, err := queries.GetAllUnresolvedIdentityPolicyPathsOnArnWithAction(s.ctx, s.db, arnString, actionName)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		// Filter through
		resolvedPaths, err := analyze.ResolveResourceAgainstIdentityPolicies(&analyze.ActionPathSet{}, identityPaths)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		if resolvedPaths == nil {
			log.Print("No paths found")
			c.IndentedJSON(http.StatusOK, nil)
		} else {
			principalsIDs := analyze.GetPrincipalNodeIDsFromActionSet(*resolvedPaths)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
			}
			principalNodes := []*graph.Node{}
			for _, id := range principalsIDs {
				node, err := analyze.GetAWSNodeByGraphID(s.ctx, s.db, id)
				if err != nil {
					c.AbortWithError(http.StatusBadRequest, err)
				}
				principalNodes = append(principalNodes, node)
			}
			c.IndentedJSON(http.StatusOK, principalNodes)
		}
	}
}

func (s *Server) GetActiveAWSConditionKeys(c *gin.Context) {
	nodes, err := queries.GetActiveAWSConditionKeys(s.ctx, s.db)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	c.IndentedJSON(http.StatusOK, nodes.Slice())
}

func DecodeArn(encodedArn string) (string, error) {
	arn, err := base64.URLEncoding.DecodeString(encodedArn)
	if err != nil {
		return "", err
	}

	arnString := string(arn[:])
	return arnString, nil
}

func (s *Server) GetAWSResource(c *gin.Context) {
	propertyName := "arn"
	encodedArn := c.Param(propertyName)

	if arnString, err := DecodeArn((encodedArn)); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	} else {
		nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, arnString, aws.UniqueArn)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		c.IndentedJSON(http.StatusOK, nodes)
	}
}

func (s *Server) GetAWSResourceActions(c *gin.Context) {
	propertyName := "arn"
	encodedArn := c.Param(propertyName)

	if arnString, err := DecodeArn((encodedArn)); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	} else {
		nodes, err := queries.GetResouceActions(s.ctx, s.db, arnString)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		c.IndentedJSON(http.StatusOK, nodes)
	}
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

func (s *Server) GetNodeShortestPath(c *gin.Context) {
	sourceNodeIdStr := c.Param("nodeid")
	destNodeIdStr := c.Param("destnodeid")

	sourceNodeId, err := strconv.Atoi(sourceNodeIdStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	destNodeId, err := strconv.Atoi(destNodeIdStr)
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

func (s *Server) GetAWSNodeTags(c *gin.Context) {
	nodeIdStr := c.Param("nodeid")
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	tags, err := queries.GetAWSNodeTags(s.ctx, s.db, graph.ID(nodeId))

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, tags.Slice())
}

func (s *Server) GetNodePermissionPath(c *gin.Context) {
	sourceNodeIdStr := c.Param("sourcenodeid")
	destNodeIdStr := c.Param("destnodeid")

	queryParams := c.Request.URL.Query()
	action := queryParams.Get("action")

	sourceNodeId, err := strconv.Atoi(sourceNodeIdStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	destNodeId, err := strconv.Atoi(destNodeIdStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	paths, err := queries.GetNodePermissionPath(s.ctx, s.db, graph.ID(sourceNodeId), graph.ID(destNodeId), action)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, paths)
}

func (s *Server) GetNodeIdentityPath(c *gin.Context) {
	sourceNodeIdStr := c.Param("nodeid")
	destNodeIdStr := c.Param("destnodeid")

	sourceNodeId, err := strconv.Atoi(sourceNodeIdStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	destNodeId, err := strconv.Atoi(destNodeIdStr)
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

func (s *Server) GetInboundRoles(c *gin.Context) {
	roleId := c.Param("roleid")

	//paths, err := queries.GetAWSRoleInboundRoleAssumptionPaths(s.ctx, s.db, roleId)
	query := "MATCH p=(a:UniqueArn) - [:IdentityTransform* {name: 'sts:assumerole'}] -> (b:AWSRole) WHERE b.roleid = '%s' AND ALL(n IN nodes(p) WHERE SINGLE(x IN nodes(p) WHERE x = n)) RETURN p"
	query = fmt.Sprintf(query, roleId)
	paths, err := queries.CypherQueryPaths(s.ctx, s.db, query)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, paths)

}

func (s *Server) GetRoleOutboundRoles(c *gin.Context) {

	roleId := c.Param("roleid")

	//paths, err := queries.GetAWSRoleInboundRoleAssumptionPaths(s.ctx, s.db, roleId)
	query := "MATCH p=(a:AWSRole) - [:IdentityTransform* {name: 'sts:assumerole'}] -> (b:AWSRole) WHERE a.roleid = '%s' AND ALL(n IN nodes(p) WHERE SINGLE(x IN nodes(p) WHERE x = n)) RETURN p"
	query = fmt.Sprintf(query, roleId)
	paths, err := queries.CypherQueryPaths(s.ctx, s.db, query)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, paths)

}

func (s *Server) GetUserOutboundRoles(c *gin.Context) {

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

func (s *Server) GetRoleRSOP(c *gin.Context) {
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

func (s *Server) GetRoleRSOPPrincipals(c *gin.Context) {
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

func (s *Server) GetRoleRSOPActions(c *gin.Context) {
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

func (s *Server) GetUserRSOPActions(c *gin.Context) {
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

func (s *Server) GetAllAssumeRoles(c *gin.Context) {
	err := queries.CreateAssumeRoleEdges(s.ctx, s.db)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.Done()
}

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

func (s *Server) GenerateStatement(c *gin.Context) {
	statementHash := c.Param("statementhash")

	statementNode, err := queries.GetAllAWSNodes(s.ctx, s.db, map[string][]string{"hash": {statementHash}})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if len(statementNode) == 0 {
		c.IndentedJSON(http.StatusOK, nil)
	}

	statementObject, err := queries.GenerateStatementObject(s.ctx, s.db, *statementNode[0])

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, statementObject)
}

func (s *Server) GetStatementPolicies(c *gin.Context) {
	statementHash := c.Param("statementhash")

	paths, err := queries.GetPoliciesAttachedToStatement(s.ctx, s.db, statementHash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusOK, paths)

}

func (s *Server) handleRequests() {
	router := gin.Default()
	router.Use(corsMiddleware())
	router.GET("/roles/:roleid", s.GetAWSRole)
	router.GET("/roles/:roleid/managedpolicies", s.GetRoleManagedPolicies)
	router.GET("/roles/:roleid/inlinepolicies", s.GetRoleInlinePolicies)
	router.GET("/roles/:roleid/inboundroles", s.GetInboundRoles)
	router.GET("/roles/:roleid/outboundroles", s.GetRoleOutboundRoles)
	router.GET("/roles/:roleid/rsop", s.GetRoleRSOP)
	router.GET("/roles/:roleid/rsop/principals", s.GetRoleRSOPPrincipals)
	router.GET("/roles/:roleid/rsop/actions", s.GetRoleRSOPActions)
	router.GET("/users/:userid", s.GetAWSUser)
	router.GET("/users/:userid/managedpolicies", s.GetUserManagedPolicies)
	router.GET("/users/:userid/inlinepolicies", s.GetUserInlinePolicies)
	router.GET("/users/:userid/rsop", s.GetUserRSOP)
	router.GET("/users/:userid/rsop/actions", s.GetUserRSOPActions)
	router.GET("/users/:userid/outboundroles", s.GetUserOutboundRoles)
	router.GET("/statements/:statementhash/generatestatement", s.GenerateStatement)
	router.GET("/statements/:statementhash/attachedpolicies", s.GetStatementPolicies)
	router.GET("/managedpolicies/:policyid", s.GetAWSPolicy)
	router.GET("/managedpolicies/:policyid/generatepolicy", s.GenerateManagedPolicy)
	router.GET("/managedpolicies/:policyid/principals", s.GetAWSPolicyPrincipals)
	router.GET("/inlinepolicies/:policyhash/generatepolicy", s.GenerateInlinePolicy)
	router.GET("/groups/:groupid", s.GetAWSGroup)
	router.GET("/groups/:groupid/members", s.GetAWSGroupMembers)
	router.GET("/groups/:groupid/policies", s.GetGroupPolicies)
	router.GET("/accounts", s.GetAWSAccountIDs)
	router.GET("/accounts/:account_id", s.GetAWSAccount)
	router.GET("/accounts/:account_id/services", s.GetAWSAccountServices)
	router.GET("/resources/:arn", s.GetAWSResource)
	router.GET("/resources/:arn/actions", s.GetAWSResourceActions)
	router.GET("/resources/:arn/inboundpermissions", s.GetAWSResourceInboundPermissions)
	router.GET("/resources/:arn/inboundpermissions/principals", s.GetAWSResourceInboundPermissionsPrincipals)
	router.GET("/resources/:arn/inboundpermissions/principals/:principalArn", s.GetActionsWithPrincipalOnResource)
	router.GET("/conditionkeys/active", s.GetActiveAWSConditionKeys)
	router.GET("/node", s.GetAWSNodes)
	router.GET("/node/:nodeid", s.GetAWSNodeByGraphID)
	router.GET("/node/:nodeid/shortestpath/:destnodeid", s.GetNodeShortestPath)
	//router.GET("/node/:nodeid/permissionpath/:destnodeid", s.GetNodePermissionPath)
	router.GET("/node/:nodeid/identitypath/:destnodeid", s.GetNodeIdentityPath)
	router.GET("/node/:nodeid/inboundedges", s.GetAWSNodeInboundEdges)
	router.GET("/node/:nodeid/outboundedges", s.GetAWSNodeOutboundEdges)
	router.GET("/node/:nodeid/tags", s.GetAWSNodeTags)
	router.GET("/permissionpath/:sourcenodeid/:destnodeid", s.GetNodePermissionPath)
	router.GET("/relationship/:relationshipid", s.GetAWSRelationshipByGraphID)
	router.GET("/analyze/assumeroles", s.GetAllAssumeRoles)
	router.GET("/search", s.Search)
	router.POST("/query", s.PostQuery)
	router.Run("0.0.0.0:4400")
}

func (s *Server) InitializeServer() {
	configFilePath := "dawgsConfig.json"
	s.ctx = context.Background()

	bhCfg, err := config.GetConfiguration(configFilePath)
	if err != nil {
		log.Fatalf("Unable to read configuration %s: %v", configFilePath, err)
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
