package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	// TODO check status of postgres and redis connections

	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}
