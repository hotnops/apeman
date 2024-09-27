package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hotnops/apeman/go/internal/queries"
	"github.com/hotnops/apeman/graphschema/aws"
)

func (s *Server) GetAWSAccount(c *gin.Context) {
	propertyName := "account_id"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSNodeByKindID(s.ctx, s.db, propertyName, id, aws.AWSAccount)
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

func (s *Server) GetAWSAccountServices(c *gin.Context) {
	propertyName := "account_id"
	id := c.Param(propertyName)

	nodes, err := queries.GetAWSAccountServices(s.ctx, s.db, id)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (s *Server) addAccountsEndpoints(router *gin.RouterGroup) {
	router.GET("", s.GetAWSAccount)
	router.GET("services", s.GetAWSAccountServices)
}
