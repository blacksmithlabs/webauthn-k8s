package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /authentication/ end point to handle getting the params for authenticating a credential
func BeginAuthentication(c *gin.Context) {
	// webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)

	// c.JSON(http.StatusOK, gin.H{"requestId": requestId, "sessionData": sessionDataJson})
	// userId := c.Param("userId")
	// c.JSON(http.StatusOK, gin.H{"userId": userId})
	c.JSON(http.StatusUnavailableForLegalReasons, gin.H{"message": "This feature is not yet implemented"})
}

// PUT /authentication/:requestId end point to handle actually authenticating a credential
func FinishAuthentication(c *gin.Context) {
	// requestId := c.Param("requestId")
	// c.JSON(http.StatusOK, gin.H{"requestId": requestId})
}
