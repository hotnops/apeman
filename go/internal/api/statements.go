package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/go/internal/queries"
)

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

func (s *Server) addStatementEndpoints(router *gin.RouterGroup) {
	router.GET("generatestatement", s.GenerateStatement)
	router.GET("/statements/:statementhash/attachedpolicies", s.GetStatementPolicies)
}
