package credential_service

import (
	"encoding/json"
	"fmt"

	"blacksmithlabs.dev/webauthn-k8s/auth/utils"
	"blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgtype"
)

type CredentialModel struct {
	webauthn.Credential
	User utils.Relationship[UserModel]
}

func CredentialModelFromDatabase(credential credentials.WebauthnCredential) (*CredentialModel, error) {
	var transport []protocol.AuthenticatorTransport
	if err := json.Unmarshal(credential.Transport, &transport); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Transport: %w", err)
	}
	var flags webauthn.CredentialFlags
	if err := json.Unmarshal(credential.Flags, &flags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Flags: %w", err)
	}
	var authenticator webauthn.Authenticator
	if err := json.Unmarshal(credential.Authenticator, &authenticator); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Authenticator: %w", err)
	}
	var attestation webauthn.CredentialAttestation
	if err := json.Unmarshal(credential.Attestation, &attestation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Attestation: %w", err)
	}

	authenticator.SignCount = uint32(credential.UseCounter)

	model := webauthn.Credential{
		ID:              credential.CredentialID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType.String,
		Transport:       transport,
		Flags:           flags,
		Authenticator:   authenticator,
		Attestation:     attestation,
	}

	return &CredentialModel{
		Credential: model,
		User: utils.Relationship[UserModel]{Loaded: false, Value: UserModel{
			ID: credential.UserID.Int64,
		}},
	}, nil
}

func (c *CredentialModel) SetUser(user *UserModel) {
	c.User.Loaded = true
	c.User.Value = *user
}

func (c *CredentialModel) ToInsertParams() (*credentials.InsertCredentialParams, error) {
	if c.User.Value.ID == 0 {
		return nil, fmt.Errorf("failed to convert: User ID is required")
	}

	transportJson, err := json.Marshal(c.Transport)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Transport: %w", err)
	}
	flagsJson, err := json.Marshal(c.Flags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Flags: %w", err)
	}
	authenticatorJson, err := json.Marshal(c.Authenticator)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Authenticator: %w", err)
	}
	attestationJson, err := json.Marshal(c.Attestation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Attestation: %w", err)
	}

	return &credentials.InsertCredentialParams{
		CredentialID:    c.ID,
		UserID:          pgtype.Int8{Int64: c.User.Value.ID, Valid: true},
		PublicKey:       c.PublicKey,
		AttestationType: pgtype.Text{String: c.AttestationType, Valid: true},
		Transport:       transportJson,
		Flags:           flagsJson,
		Authenticator:   authenticatorJson,
		Attestation:     attestationJson,
	}, nil
}
