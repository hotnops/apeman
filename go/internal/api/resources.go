package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/analyze"
	"github.com/hotnops/apeman/go/internal/queries"
	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/specterops/bloodhound/dawgs/graph"
)

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

func (s *Server) addResourceEndpoints(router *gin.RouterGroup) {
	router.GET("", s.GetAWSResource)
	router.GET("actions", s.GetAWSResourceActions)
	router.GET("inboundpermissions", s.GetAWSResourceInboundPermissions)
	router.GET("inboundpermissions/principals", s.GetAWSResourceInboundPermissionsPrincipals)
	router.GET("inboundpermissions/principals/:principalArn", s.GetActionsWithPrincipalOnResource)
}
