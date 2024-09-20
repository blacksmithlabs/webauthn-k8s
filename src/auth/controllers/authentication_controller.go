package controllers

import (
	"net/http"

	credential_service "blacksmithlabs.dev/webauthn-k8s/auth/services/credential"
	"blacksmithlabs.dev/webauthn-k8s/auth/services/request_cache"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// POST /authentication/ end point to handle getting the params for authenticating a credential
func BeginAuthentication(c *gin.Context) {
	var requestPayload dto.StartAuthenticationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}
	if err := requestPayload.Validate(); err != nil {
		logger.Error("Invalid request payload", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request payload"})
		return
	}

	service, err := credential_service.New(c)
	if err != nil {
		logger.Error("Failed to get credentials service", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Database error"})
		return
	}
	user, err := service.GetUserWithCredentialsByRef(requestPayload.User.UserID, false)
	if err != nil {
		logger.Error("Failed to get user", "error", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error(), "message": "User not found"})
		return
	}
	if !user.Credentials.Loaded || len(user.Credentials.Value) == 0 {
		logger.Error("User has no credentials", "user", user)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No credentials found", "message": "User has no credentials"})
		return
	}

	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	options, sessionData, err := webAuthn.BeginLogin(user)
	if err != nil {
		logger.Error("Failed to create authentication options", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to create authentication options"})
		return
	}

	requestId := uuid.New().String()

	cache := request_cache.New(c)
	requestInfo := request_cache.RequestInfo{UserId: user.ID, SessionData: sessionData}
	if err := cache.SetRequestCache(requestId, &requestInfo); err != nil {
		logger.Error("Failed to save request data to cache", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to save request data to cache"})
		return
	}

	c.JSON(http.StatusOK, dto.StartAuthenticationResponse{
		RequestID: requestId,
		Options:   *options,
	})
}

// PUT /authentication/:requestId end point to handle actually authenticating a credential
func FinishAuthentication(c *gin.Context) {
	requestId := c.Param("requestId")
	logger.Info("Finish authentication", "requestId", requestId)

	cache := request_cache.New(c)
	requestInfo, err := cache.GetRequestCache(requestId)
	if err == cache.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Request Not Found", "requestId": requestId})
		return
	} else if err != nil {
		logger.Error("Failed to get request data from cache", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to get request data from cache"})
		return
	}

	userId := requestInfo.UserId
	sessionData := requestInfo.SessionData

	// Get the user for this credential
	service, err := credential_service.New(c)
	if err != nil {
		logger.Error("Failed to get credentials service", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Database error"})
		return
	}

	user, err := service.GetUserWithCredentialsByID(userId, false)
	if err != nil {
		logger.Error("Failed to get user", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "User lookup failed"})
		return
	}

	logger.Info("Loaded user for request", "user", user)

	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	var requestPayload dto.FinishAuthenticationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}

	logger.Info("Preparing to parse assertion", "assertion", requestPayload.Assertion)

	parsedAssertion, err := requestPayload.Assertion.Parse()
	if err != nil {
		logger.Error("Failed to parse assertion", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err, "message": "Failed to parse assertion"})
		return
	}

	credential, err := webAuthn.ValidateLogin(user, *sessionData, parsedAssertion)
	if err != nil {
		logger.Error("Failed to validate login", "error", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err, "message": "Failed to validate login"})
		return
	}

	count, err := service.IncrementCredentialUseCounter(credential.ID)
	if err != nil {
		logger.Error("Failed to increment credential use counter", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to increment credential use counter"})
		return
	}

	// Clear request cache since request is finished
	cache.DeleteRequestCache(requestId)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated", "credential": credential, "useCount": count})
}
