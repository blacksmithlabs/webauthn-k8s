package dto

import (
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// RegistrationUserInfo is a struct that holds the user information for a credential.
type RegistrationUserInfo struct {
	UserId      string `json:"userId" binding:"required"`
	UserName    string `json:"userName" binding:"required"`
	DisplayName string `json:"displayName"`
}

func (r RegistrationUserInfo) WebAuthnID() []byte {
	return []byte(r.UserId)
}

func (r RegistrationUserInfo) WebAuthnName() string {
	return r.UserName
}

func (r RegistrationUserInfo) WebAuthnDisplayName() string {
	if r.DisplayName == "" {
		return r.UserName
	}
	return r.DisplayName
}

// There are no credentials for the user
func (r RegistrationUserInfo) WebAuthnCredentials() []webauthn.Credential {
	return []webauthn.Credential{}
}

// Validate validates the RegistrationUserInfo.
func (r RegistrationUserInfo) Validate() error {
	if r.UserId == "" {
		return fmt.Errorf("userId is required")
	}
	if r.UserName == "" {
		return fmt.Errorf("userName is required")
	}
	return nil
}

// CreateRegistrationRequest is a struct that holds the request for creating a credential.
type CreateRegistrationRequest struct {
	User RegistrationUserInfo `json:"user"`
}

// Validate validates the CreateRegistrationRequest.
func (c CreateRegistrationRequest) Validate() error {
	return c.User.Validate()
}

// CreateRegistrationResponse is a struct that holds the response for creating a credential.
type CreateRegistrationResponse struct {
	RequestId string                      `json:"requestId"`
	Options   protocol.CredentialCreation `json:"options"`
}
