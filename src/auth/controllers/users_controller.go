package controllers

import (
	"net/http"

	credential_service "blacksmithlabs.dev/webauthn-k8s/auth/services/credential"
	"blacksmithlabs.dev/webauthn-k8s/auth/utils"
	"github.com/gin-gonic/gin"
)

type ResponseCredentials struct {
	// The Credential ID of the public key credential source
	ID []byte `json:"id"`

	// The credential public key of the public key credential source
	PublicKey []byte `json:"publicKey"`
}

// GET /users/:userId/credentials end point to handle getting the credentials for a user
func GetUserCredentials(c *gin.Context) {
	userId := c.Param("userId")

	service, err := credential_service.New(c)
	if err != nil {
		logger.Error("Failed to get credentials service", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Database error"})
		return
	}

	user, err := service.GetUserWithCredentialsByRef(userId, true)
	if err != nil {
		logger.Error("Failed to get user", "error", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error(), "message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credentials": utils.Map(user.Credentials.Value, func(c credential_service.CredentialModel) ResponseCredentials {
		return ResponseCredentials{
			ID:        c.ID,
			PublicKey: c.PublicKey,
		}
	})})
}
