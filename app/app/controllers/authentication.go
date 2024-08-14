package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BeginAuthentication(c *gin.Context) {
	if len(c.Request.Cookies()) == 0 {
		logger.Info("No cookies found")
	} else {
		for _, cookie := range c.Request.Cookies() {
			logger.Info("Cookie", "name", cookie.Name, "value", cookie.Value)
		}
	}

	session := getSession(c)
	requestId := session.Get("requestId")
	if requestId == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No request ID found in session"})
		return
	}

	sessionDataJson := session.Get(requestId).(string)

	c.JSON(http.StatusOK, gin.H{"requestId": requestId, "sessionData": sessionDataJson})
	// userId := c.Param("userId")
	// c.JSON(http.StatusOK, gin.H{"userId": userId})
}

func FinishAuthentication(c *gin.Context) {
	// requestId := c.Param("requestId")
	// c.JSON(http.StatusOK, gin.H{"requestId": requestId})
}
