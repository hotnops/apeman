package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/src/api/src/queries"
)

func (s *Server) GetActionPolicies(c *gin.Context) {
	actionName := "actionname"
	action := c.Param(actionName)

	statements, err := queries.GetActionPolicies(s.ctx, s.db, action)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, statements)
}

func (s *Server) addActionsEndpoints(router *gin.RouterGroup) {
	router.GET("policies", s.GetActionPolicies)
}
