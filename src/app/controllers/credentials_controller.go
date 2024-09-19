package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	credential_service "blacksmithlabs.dev/webauthn-k8s/app/services/credential"
	"blacksmithlabs.dev/webauthn-k8s/app/services/request_cache"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
)

// POST /credentials/ end point to handle getting the params for creating a credential
func BeginCreateCredential(c *gin.Context) {
	var requestPayload dto.StartRegistrationRequest
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

	// Upsert the user for this credential
	service, err := credential_service.New(c)
	if err != nil {
		logger.Error("Failed to get credentials service", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Database error"})
		return
	}

	user, err := service.UpsertUser(requestPayload.User)
	if err != nil {
		logger.Error("Failed to upsert user", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "User creation failed"})
		return
	} else {
		logger.Info("User upserted", "user", user)
	}

	logger.Info("Creating credential for user", "userId", user.ID, "refId", user.RefID)

	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	options, sessionData, err := webAuthn.BeginRegistration(user)
	if err != nil {
		logger.Error("Failed to create registration options", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to create registration options"})
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

	c.JSON(http.StatusOK, dto.StartRegistrationResponse{
		RequestID: requestId,
		Options:   *options,
	})
}

// PUT /credentials/:requestId end point to handle actually creating a credential
func FinishCreateCredential(c *gin.Context) {
	requestId := c.Param("requestId")
	logger.Info("Finish create credential", "requestId", requestId)

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

	user, err := service.GetUserByID(userId)
	if err != nil {
		logger.Error("Failed to get user", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "User lookup failed"})
		return
	}

	logger.Info("Loaded user for request", "user", user)

	var requestPayload dto.FinishRegistrationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}
	parsedCredential, err := requestPayload.Credential.Parse()
	if err != nil {
		logger.Error("Failed to parse credential", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err, "message": "Failed to parse credential"})
		return
	}

	// Step 1 - 16
	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	credential, err := webAuthn.CreateCredential(user, *sessionData, parsedCredential)
	if err != nil {
		logger.Error("Failed to finish registration", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to finish registration"})
		return
	}

	logger.Info("Credential created", "credential", credential)

	// Step 17 - Check that the credentialId is not yet registered to any other user
	// Step 18 - Associate the credential with the user account
	err = service.InsertCredential(user, credential)
	if err != nil {
		logger.Error("Failed to insert credential", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"erqror": err.Error(), "message": "Failed to insert credential"})
		return
	}

	// Clear request cache since request is finished
	cache.DeleteRequestCache(requestId)

	c.JSON(http.StatusOK, dto.FinishRegistrationResponse{
		RequestID:  requestId,
		Credential: dto.CredentialResponseFromWebauthn(credential),
	})
}
