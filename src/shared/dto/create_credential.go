package dto

import (
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// RegistrationUserInfo is a struct that holds the user information for a credential.
type RegistrationUserInfo struct {
	UserID      string `json:"userId" binding:"required"`
	UserName    string `json:"userName"`
	DisplayName string `json:"displayName"`
}

// Validate validates the RegistrationUserInfo.
func (r RegistrationUserInfo) Validate() error {
	if r.UserID == "" {
		return fmt.Errorf("userId is required")
	}
	if r.UserName == "" {
		return fmt.Errorf("userName is required")
	}
	return nil
}

// StartRegistrationRequest is a struct that holds the request for creating a credential.
type StartRegistrationRequest struct {
	User RegistrationUserInfo `json:"user" binding:"required"`
}

// Validate validates the StartRegistrationRequest.
func (c StartRegistrationRequest) Validate() error {
	return c.User.Validate()
}

// StartRegistrationResponse is a struct that holds the response for creating a credential.
type StartRegistrationResponse struct {
	RequestID string                      `json:"requestId" binding:"required"`
	Options   protocol.CredentialCreation `json:"options" binding:"required"`
}

// FinishRegistrationRequest is a struct that holds the request for finishing a credential.
type FinishRegistrationRequest struct {
	User       RegistrationUserInfo                `json:"user"`
	Credential protocol.CredentialCreationResponse `json:"credential" binding:"required"`
}

type CredentialResponse struct {
	ID        protocol.URLEncodedBase64 `json:"id" binding:"required"`
	PublicKey protocol.URLEncodedBase64 `json:"publicKey" binding:"required"`
}

func CredentialResponseFromWebauthn(credential *webauthn.Credential) CredentialResponse {
	return CredentialResponse{
		ID:        protocol.URLEncodedBase64(credential.ID),
		PublicKey: protocol.URLEncodedBase64(credential.PublicKey),
	}
}

type FinishRegistrationResponse struct {
	RequestID  string             `json:"requestId" binding:"required"`
	Credential CredentialResponse `json:"credential" binding:"required"`
}
