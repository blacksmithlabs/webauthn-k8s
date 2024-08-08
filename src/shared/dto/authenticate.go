package dto

import (
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
)

type AuthenticationUserInfo struct {
	UserID string `json:"userId" binding:"required"`
}

func (a AuthenticationUserInfo) Validate() error {
	if a.UserID == "" {
		return fmt.Errorf("userId is required")
	}
	return nil
}

// StartAuthenticationRequest is a struct that holds the request for starting authentication.
type StartAuthenticationRequest struct {
	User AuthenticationUserInfo `json:"user" binding:"required"`
}

// Validate validates the StartAuthenticationRequest.
func (c StartAuthenticationRequest) Validate() error {
	return c.User.Validate()
}

// Response for starting authentication
type StartAuthenticationResponse struct {
	RequestID string                       `json:"requestId" binding:"required"`
	Options   protocol.CredentialAssertion `json:"options" binding:"required"`
}

// FinishAuthenticationRequest is a struct that holds the request for finishing authentication.
type FinishAuthenticationRequest struct {
	User      AuthenticationUserInfo               `json:"user"`
	Assertion protocol.CredentialAssertionResponse `json:"assertion" binding:"required"`
}
