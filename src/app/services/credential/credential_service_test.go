package credential_service

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
	"blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/milqa/pgxpoolmock"
	"github.com/milqa/pgxpoolmock/sqlc"
)

var mockPool *pgxpoolmock.MockPgxIface

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func setupTest(t *testing.T) {
	oldGetDbConn := getDbConn

	ctrl := gomock.NewController(t)

	mockPool = pgxpoolmock.NewMockPgxIface(ctrl)
	getDbConn = func(ctx context.Context) (sqlc.DBTX, error) {
		return mockPool, nil
	}

	t.Cleanup(func() {
		getDbConn = oldGetDbConn
		ctrl.Finish()
	})
}

func TestNew(t *testing.T) {
	setupTest(t)

	_, err := New(context.Background())
	if err != nil {
		t.Errorf("New() error = %v, want nil", err)
	}
}

func TestCredentialService_UpsertUser(t *testing.T) {
	// Given
	setupTest(t)

	mockPool.EXPECT().QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:INSERT INTO webauthn_users.*)"),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(
		pgxpoolmock.NewRow(int64(1), "123", []byte("123"), "User Name", "Display Name"),
	)

	expected := UserModel{
		ID:          1,
		RefID:       "123",
		RawID:       []byte("123"),
		Name:        "User Name",
		DisplayName: "Display Name",
	}

	// When
	credentialService, err := New(context.Background())
	if err != nil {
		t.Errorf("New() error = %v, want nil", err)
	}

	user, err := credentialService.UpsertUser(dto.RegistrationUserInfo{
		UserID:      "123",
		UserName:    "User Name",
		DisplayName: "Display Name",
	})

	// Then
	if err != nil {
		t.Errorf("UpsertUsers() error = %v, want nil", err)
	}
	if user == nil {
		t.Errorf("UpsertUsers() user = nil, want not nil")
	} else if !reflect.DeepEqual(*user, expected) {
		t.Errorf("UpsertUsers() user = %v, want %v", *user, expected)
	}
}

func TestCredentialService_GetUserByID(t *testing.T) {
	// Given
	setupTest(t)

	// Mock query
	mockPool.EXPECT().QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?s:.*SELECT.*FROM webauthn_users.*WHERE _id =.*)"),
		int64(2),
	).Return(
		pgxpoolmock.NewRow(int64(2), "ref-id", []byte("ref-id"), "Name", "DisplayName"),
	)

	expected := UserModel{
		ID:          2,
		RefID:       "ref-id",
		RawID:       []byte("ref-id"),
		Name:        "Name",
		DisplayName: "DisplayName",
	}

	// When
	credentialService, err := New(context.Background())
	if err != nil {
		t.Errorf("New() error = %v, want nil", err)
	}

	user, err := credentialService.GetUserByID(2)

	// Then
	if err != nil {
		t.Errorf("GetUserByID() error = %v, want nil", err)
	}
	if user == nil {
		t.Errorf("GetUserByID() user = nil, want not nil")
	} else if !reflect.DeepEqual(*user, expected) {
		t.Errorf("GetUserByID() user = %v, want %v", *user, expected)
	}
}

func TestCredentialService_GetUserByRef(t *testing.T) {
	// Given
	setupTest(t)

	// Mock query
	mockPool.EXPECT().QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?s:.*SELECT.*FROM webauthn_users.*WHERE ref_id =.*)"),
		"ref-id",
	).Return(
		pgxpoolmock.NewRow(int64(2), "ref-id", []byte("ref-id"), "Name", "DisplayName"),
	)

	expected := UserModel{
		ID:          2,
		RefID:       "ref-id",
		RawID:       []byte("ref-id"),
		Name:        "Name",
		DisplayName: "DisplayName",
	}

	// When
	credentialService, err := New(context.Background())
	if err != nil {
		t.Errorf("New() error = %v, want nil", err)
	}

	user, err := credentialService.GetUserByRef("ref-id")

	// Then
	if err != nil {
		t.Errorf("GetUserByID() error = %v, want nil", err)
	}
	if user == nil {
		t.Errorf("GetUserByID() user = nil, want not nil")
	} else if !reflect.DeepEqual(*user, expected) {
		t.Errorf("GetUserByID() user = %v, want %v", *user, expected)
	}
}

func Test_parseUserCredentialList(t *testing.T) {
	type args struct {
		userCredentials []credentials.ResultUserCredentialRow
	}

	testUser := credentials.WebauthnUser{
		ID:          1,
		RefID:       "ref-id",
		RawID:       []byte("ref-id"),
		Name:        "Name",
		DisplayName: "DisplayName",
	}
	testCredential1 := credentials.WebauthnCredential{
		CredentialID:    []byte("cred-id-1"),
		UserID:          pgtype.Int8{Int64: testUser.ID, Valid: true},
		UseCounter:      0,
		PublicKey:       []byte("public-key-1"),
		AttestationType: pgtype.Text{String: "attestation-type-1", Valid: true},
		Transport:       []byte("[\"usb\",\"nfc\"]"),
		Flags:           []byte("{\"userPresent\":true,\"userVerified\":false}"),
		Authenticator:   []byte(fmt.Sprintf("{\"AAGUID\":\"%s\",\"signCount\":0,\"attachment\":\"platform\"}", b64("aaguid-1"))),
		Attestation: []byte(fmt.Sprintf(
			"{\"clientDataJSON\":\"%s\",\"clientDataHash\":\"%s\",\"authenticatorData\":\"%s\",\"publicKeyAlgorithm\":-7,\"object\":\"%s\"}",
			b64("client-data-json-1"),
			b64("client-data-hash-1"),
			b64("authenticator-data-1"),
			b64("attestation-object-1"),
		)),
	}
	testCredential2 := credentials.WebauthnCredential{
		CredentialID:    []byte("cred-id-2"),
		UserID:          pgtype.Int8{Int64: testUser.ID, Valid: true},
		UseCounter:      1,
		PublicKey:       []byte("public-key-2"),
		AttestationType: pgtype.Text{String: "attestation-type-2", Valid: true},
		Transport:       []byte("[\"ble\",\"hybrid\"]"),
		Flags:           []byte("{\"userPresent\":true,\"userVerified\":true}"),
		Authenticator:   []byte(fmt.Sprintf("{\"AAGUID\":\"%s\",\"signCount\":0,\"attachment\":\"cross-platform\"}", b64("aaguid-2"))),
		Attestation: []byte(fmt.Sprintf(
			"{\"clientDataJSON\":\"%s\",\"clientDataHash\":\"%s\",\"authenticatorData\":null,\"publicKeyAlgorithm\":-7,\"object\":\"%s\"}",
			b64("client-data-json-2"),
			b64("client-data-hash-2"),
			b64("attestation-object-2"),
		)),
	}

	expectedUser := &UserModel{
		ID:          testUser.ID,
		RefID:       testUser.RefID,
		RawID:       testUser.RawID,
		Name:        testUser.Name,
		DisplayName: testUser.DisplayName,
	}
	expectedUser.addCredential(webauthn.Credential{
		ID:              []byte("cred-id-1"),
		PublicKey:       []byte("public-key-1"),
		AttestationType: "attestation-type-1",
		Transport:       []protocol.AuthenticatorTransport{protocol.USB, protocol.NFC},
		Flags:           webauthn.CredentialFlags{UserPresent: true, UserVerified: false},
		Authenticator:   webauthn.Authenticator{AAGUID: []byte("aaguid-1"), SignCount: 0, Attachment: protocol.Platform},
		Attestation: webauthn.CredentialAttestation{
			ClientDataJSON:     []byte("client-data-json-1"),
			ClientDataHash:     []byte("client-data-hash-1"),
			AuthenticatorData:  []byte("authenticator-data-1"),
			PublicKeyAlgorithm: -7,
			Object:             []byte("attestation-object-1"),
		},
	})
	expectedUser.addCredential(webauthn.Credential{
		ID:              []byte("cred-id-2"),
		PublicKey:       []byte("public-key-2"),
		AttestationType: "attestation-type-2",
		Transport:       []protocol.AuthenticatorTransport{protocol.BLE, protocol.Hybrid},
		Flags:           webauthn.CredentialFlags{UserPresent: true, UserVerified: true},
		Authenticator:   webauthn.Authenticator{AAGUID: []byte("aaguid-2"), SignCount: 1, Attachment: protocol.CrossPlatform},
		Attestation: webauthn.CredentialAttestation{
			ClientDataJSON:     []byte("client-data-json-2"),
			ClientDataHash:     []byte("client-data-hash-2"),
			PublicKeyAlgorithm: -7,
			Object:             []byte("attestation-object-2"),
		},
	})

	tests := []struct {
		name    string
		args    args
		want    *UserModel
		wantErr bool
	}{
		{
			name:    "Empty list",
			args:    args{},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Credentials By ID",
			args: args{
				userCredentials: []credentials.ResultUserCredentialRow{
					credentials.GetUserWithCredentialsByIDRow{
						WebauthnUser:       testUser,
						WebauthnCredential: testCredential1,
					},
					credentials.GetUserWithCredentialsByIDRow{
						WebauthnUser:       testUser,
						WebauthnCredential: testCredential2,
					},
				},
			},
			want:    expectedUser,
			wantErr: false,
		},
		{
			name: "Credentials By Ref",
			args: args{
				userCredentials: []credentials.ResultUserCredentialRow{
					credentials.GetUserWithCredentialsByRefRow{
						WebauthnUser:       testUser,
						WebauthnCredential: testCredential1,
					},
					credentials.GetUserWithCredentialsByRefRow{
						WebauthnUser:       testUser,
						WebauthnCredential: testCredential2,
					},
				},
			},
			want:    expectedUser,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUserCredentialList(tt.args.userCredentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUserCredentialList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseUserCredentialList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialService_GetUserWithCredentialsByID(t *testing.T) {
	type fields struct {
		ctx     context.Context
		conn    sqlc.DBTX
		queries *credentials.Queries
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *UserModel
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CredentialService{
				ctx:     tt.fields.ctx,
				conn:    tt.fields.conn,
				queries: tt.fields.queries,
			}
			got, err := s.GetUserWithCredentialsByID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("CredentialService.GetUserWithCredentialsByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CredentialService.GetUserWithCredentialsByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialService_GetUserWithCredentialsByRef(t *testing.T) {
	type fields struct {
		ctx     context.Context
		conn    sqlc.DBTX
		queries *credentials.Queries
	}
	type args struct {
		ref string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *UserModel
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CredentialService{
				ctx:     tt.fields.ctx,
				conn:    tt.fields.conn,
				queries: tt.fields.queries,
			}
			got, err := s.GetUserWithCredentialsByRef(tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("CredentialService.GetUserWithCredentialsByRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CredentialService.GetUserWithCredentialsByRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialService_InsertCredential(t *testing.T) {
	type fields struct {
		ctx     context.Context
		conn    sqlc.DBTX
		queries *credentials.Queries
	}
	type args struct {
		user       *UserModel
		credential *webauthn.Credential
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CredentialService{
				ctx:     tt.fields.ctx,
				conn:    tt.fields.conn,
				queries: tt.fields.queries,
			}
			if err := s.InsertCredential(tt.args.user, tt.args.credential); (err != nil) != tt.wantErr {
				t.Errorf("CredentialService.InsertCredential() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCredentialService_IncrementCredentialUseCounter(t *testing.T) {
	type fields struct {
		ctx     context.Context
		conn    sqlc.DBTX
		queries *credentials.Queries
	}
	type args struct {
		credentialID []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int32
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CredentialService{
				ctx:     tt.fields.ctx,
				conn:    tt.fields.conn,
				queries: tt.fields.queries,
			}
			got, err := s.IncrementCredentialUseCounter(tt.args.credentialID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CredentialService.IncrementCredentialUseCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CredentialService.IncrementCredentialUseCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}
