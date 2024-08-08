package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	// TODO check status of postgres and redis connections

	// TODO remove this after load balance testing
	hostname, err := os.Hostname()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get hostname",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "OK",
		"hostname": hostname,
	})
}
