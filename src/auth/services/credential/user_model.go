package credential_service

import (
	"blacksmithlabs.dev/webauthn-k8s/auth/utils"
	"blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgtype"
)

type CredentialRelationship = utils.Relationship[[]CredentialModel]

type UserModel struct {
	ID          int64
	RefID       string
	RawID       []byte
	Name        string
	DisplayName string
	Credentials CredentialRelationship
}

func (u *UserModel) PgID() pgtype.Int8 {
	return pgtype.Int8{Int64: u.ID, Valid: true}
}

func (u *UserModel) WebAuthnID() []byte {
	return u.RawID
}

func (u *UserModel) WebAuthnName() string {
	return u.Name
}

func (u *UserModel) WebAuthnDisplayName() string {
	if u.DisplayName == "" {
		return u.Name
	}
	return u.DisplayName
}

func (u *UserModel) WebAuthnCredentials() []webauthn.Credential {
	if u == nil || !u.Credentials.Loaded {
		return []webauthn.Credential{}
	}

	return utils.Map(u.Credentials.Value, func(c CredentialModel) webauthn.Credential {
		return c.Credential
	})
}

func (u *UserModel) linkCredential(credential CredentialModel) {
	if credential.User.Value.ID == 0 || credential.User.Value.ID == u.ID {
		u.Credentials.Loaded = true

		credential.User = UserRelationship{Loaded: true, Value: *u}
		u.Credentials.Value = append(u.Credentials.Value, credential)
	}
}

func UserModelFromDatabase(user credentials.WebauthnUser) *UserModel {
	return &UserModel{
		ID:          user.ID,
		RefID:       user.RefID,
		RawID:       user.RawID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
	}
}

func (u *UserModel) ToDatabase() *credentials.WebauthnUser {
	return &credentials.WebauthnUser{
		ID:          u.ID,
		RefID:       u.RefID,
		RawID:       u.RawID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
	}
}
