package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"blacksmithlabs.dev/webauthn-k8s/app/utils"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
)

var logger = utils.GetLogger()

func BeginCreateCredential(c *gin.Context) {
	var requestPayload dto.CreateRegistrationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}
	if err := requestPayload.Validate(); err != nil {
		logger.Error("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request payload"})
		return
	}

	logger.Info("Creating credential for user", "userId", requestPayload.User.UserId)

	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	options, sessionData, err := webAuthn.BeginRegistration(requestPayload.User)

	if err != nil {
		logger.Error("Failed to create registration options", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to create registration options"})
		return
	}

	requestId := uuid.New().String()
	session := getSession(c)
	// TODO move user data into the database
	logger.Info("Saving request data to session", "requestId", requestId, "user", requestPayload.User, "session", sessionData)
	session.Set(requestId, gin.H{"user": requestPayload.User, "session": sessionData})
	if err := session.Save(); err != nil {
		logger.Error("Failed to save session data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to save session data"})
		return
	}

	c.JSON(http.StatusOK, dto.CreateRegistrationResponse{
		RequestId: requestId,
		Options:   *options,
	})
}

func FinishCreateCredential(c *gin.Context) {
	requestId := c.Param("requestId")
	logger.Info("Finish create credential", "requestId", requestId)

	session := getSession(c)
	sessionPayload := session.Get(requestId)

	if sessionPayload == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No request data found for request ID", "requestId": requestId})
		return
	}

	user := sessionPayload.(gin.H)["user"].(dto.RegistrationUserInfo)
	sessionData := sessionPayload.(gin.H)["session"].(webauthn.SessionData)

	logger.Info("Loaded user for request", "user", user)

	var requestPayload dto.FinishRegistrationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}
	parsedCredential, err := requestPayload.Credential.Parse()
	if err != nil {
		logger.Error("Failed to parse credential", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err, "message": "Failed to parse credential"})
		return
	}

	// Step 1 - 16
	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	credential, err := webAuthn.CreateCredential(user, sessionData, parsedCredential)
	if err != nil {
		logger.Error("Failed to finish registration", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to finish registration"})
		return
	}

	// Step 17 - Check that the credentialId is not yet registered to any other user
	// Step 18 - Associate the credential with the user account
	// TODO save credential to database

	c.JSON(http.StatusOK, dto.FinishRegistrationResponse{
		RequestId:  requestId,
		Credential: dto.CredentialResponseFromWebauthn(credential),
	})
}
