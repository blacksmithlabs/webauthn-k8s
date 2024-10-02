package credential_service

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"blacksmithlabs.dev/webauthn-k8s/auth/database"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/milqa/pgxpoolmock"
)

var mockPool *pgxpoolmock.MockPgxIface

const getUserByIdSql = "(?s:.*SELECT.*FROM webauthn_users.*WHERE _id =.*)"
const getUserByRefSql = "(?s:.*SELECT.*FROM webauthn_users.*WHERE ref_id =.*)"

var credentialRows = []string{"credential_id", "user_id", "use_counter", "public_key", "attestation_type", "transport", "flags", "authenticator", "attestation", "meta"}

func mockCredentialRow(
	credentialId string,
	active bool,
	nickname string,
) ([]byte, pgtype.Int8, int32, []byte, pgtype.Text, []byte, []byte, []byte, []byte, []byte) {
	status := CredentialStatusActive
	if !active {
		status = CredentialStatusDisabled
	}
	return []byte(credentialId), // credential_id
		pgtype.Int8{Int64: 1, Valid: true}, // user_id
		int32(0), // use_counter
		[]byte{}, // public_key
		pgtype.Text{String: "none", Valid: true}, // attestation_type
		[]byte("[\"internal\"]"), // transports
		[]byte("{}"), // flags
		[]byte("{}"), // authenticator
		[]byte("{}"), // attestation
		[]byte(fmt.Sprintf(`{"status": "%v", "nickname": "%v"}`, status, nickname)) // meta
}

func buildUserModel(
	id int64,
	refId string,
	name string,
	displayName string,
	models ...CredentialModel,
) *UserModel {
	user := UserModel{
		ID:          id,
		RefID:       refId,
		RawID:       []byte(refId),
		Name:        name,
		DisplayName: displayName,
		Credentials: CredentialRelationship{
			Loaded: false,
		},
	}

	if len(models) > 0 {
		user.Credentials.Loaded = true
		for _, model := range models {
			model.User = UserRelationship{Loaded: true, Value: user}
			user.Credentials.Value = append(user.Credentials.Value, model)
		}
	}

	return &user
}

func buildCredentialModel(
	credentialId string,
	active bool,
	nickname string,
) CredentialModel {
	status := CredentialStatusActive
	if !active {
		status = CredentialStatusDisabled
	}
	return CredentialModel{
		Credential: webauthn.Credential{
			ID:              []byte(credentialId),
			PublicKey:       []byte{},
			AttestationType: "none",
			Transport:       []protocol.AuthenticatorTransport{"internal"},
			Flags:           webauthn.CredentialFlags{},
			Authenticator:   webauthn.Authenticator{},
			Attestation:     webauthn.CredentialAttestation{},
		},
		User: UserRelationship{Loaded: false, Value: UserModel{ID: 1}},
		Meta: CredentialMeta{
			Status:   status,
			Nickname: nickname,
		},
	}
}

func buildWebAuthnCredential(id string) *webauthn.Credential {
	return &webauthn.Credential{
		ID:              []byte(id),
		PublicKey:       []byte("public-key"),
		AttestationType: "none",
		Transport:       []protocol.AuthenticatorTransport{"internal"},
		Flags:           webauthn.CredentialFlags{},
		Authenticator:   webauthn.Authenticator{},
		Attestation:     webauthn.CredentialAttestation{},
	}
}

func setupTest(t *testing.T) {
	oldGetDbConn := getDbConn

	ctrl := gomock.NewController(t)

	mockPool = pgxpoolmock.NewMockPgxIface(ctrl)
	getDbConn = func(ctx context.Context) (database.DBConn, error) {
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

func TestCredentialService_UpsertUser_UserNotFound(t *testing.T) {
	// Given
	setupTest(t)

	mocker := mockPool.EXPECT()
	mocker.Begin(gomock.Any()).Return(mockPool, nil)
	mocker.Commit(gomock.Any()).Return(nil)
	mocker.Rollback(gomock.Any()).Return(nil)
	mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_users.*)"), "123").Return(
		pgxpoolmock.NewRow(int64(0), "", []byte{}, "", "").WithError(pgx.ErrNoRows),
	)
	mocker.QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:INSERT INTO webauthn_users.*)"),
		"123",
		gomock.Any(), // Random raw ID
		"User Name",
		"Display Name",
	).Return(
		pgxpoolmock.NewRow(int64(1), "123", []byte("123"), "User Name", "Display Name"),
	)

	expected := buildUserModel(1, "123", "User Name", "Display Name")

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
	} else if !reflect.DeepEqual(user, expected) {
		t.Errorf("UpsertUsers() user = %v, want %v", user, expected)
	}
}

func TestCredentialService_UpsertUser_UserFound(t *testing.T) {
	// Given
	setupTest(t)

	mocker := mockPool.EXPECT()
	mocker.Begin(gomock.Any()).Return(mockPool, nil)
	mocker.Commit(gomock.Any()).Return(nil)
	mocker.Rollback(gomock.Any()).Return(nil)
	mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_users.*)"), "123").Return(
		pgxpoolmock.NewRow(int64(1), "123", []byte("123"), "User Name", "Display Name"),
	)
	// Insert query should not be called

	expected := buildUserModel(1, "123", "User Name", "Display Name")

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
	} else if !reflect.DeepEqual(user, expected) {
		t.Errorf("UpsertUsers() user = %v, want %v", user, expected)
	}
}

func TestCredentialService_UpsertUser_Error(t *testing.T) {
	// Given
	setupTest(t)

	mocker := mockPool.EXPECT()
	mocker.Begin(gomock.Any()).Return(mockPool, nil)
	// Commit should not be called
	mocker.Rollback(gomock.Any()).Return(nil)
	mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_users.*)"), "123").Return(
		pgxpoolmock.NewRow(int64(0), "", []byte{}, "", "").WithError(pgx.ErrNoRows),
	)
	mocker.QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:INSERT INTO webauthn_users.*)"),
		"123",
		gomock.Any(), // Random raw ID
		"User Name",
		"Display Name",
	).Return(
		pgxpoolmock.NewRow(int64(0), "", []byte{}, "", "").WithError(pgx.ErrNoRows),
	)

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
	if err == nil {
		t.Errorf("UpsertUsers() error = nil, want not nil. User: %v", user)
	}
	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("UpsertUsers() error = %v, want %v", err, "query failed")
	}
}

func TestCredentialService_GetUserByID(t *testing.T) {
	// Given
	setupTest(t)

	// Mock query
	mockPool.EXPECT().QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByIdSql), int64(2)).Return(
		pgxpoolmock.NewRow(int64(2), "ref-id", []byte("ref-id"), "Name", "DisplayName"),
	)

	expected := buildUserModel(2, "ref-id", "Name", "DisplayName")

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
	} else if !reflect.DeepEqual(user, expected) {
		t.Errorf("GetUserByID() user = %v, want %v", user, expected)
	}
}

func TestCredentialService_GetUserByRef(t *testing.T) {
	// Given
	setupTest(t)

	// Mock query
	mockPool.EXPECT().QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByRefSql), "ref-id").Return(
		pgxpoolmock.NewRow(int64(2), "ref-id", []byte("ref-id"), "Name", "DisplayName"),
	)

	expected := buildUserModel(2, "ref-id", "Name", "DisplayName")

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
	} else if !reflect.DeepEqual(user, expected) {
		t.Errorf("GetUserByID() user = %v, want %v", user, expected)
	}
}

func TestCredentialService_GetUserWithCredentialsByID(t *testing.T) {
	type setup func()
	type args struct {
		id  int64
		all bool
	}
	tests := []struct {
		name    string
		setup   setup
		args    args
		want    *UserModel
		wantErr bool
	}{
		{
			name: "User not found",
			setup: func() {
				mockPool.EXPECT().QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByIdSql), int64(1)).Return(
					pgxpoolmock.NewRow(int64(0), "", []byte{}, "", "").WithError(pgx.ErrNoRows),
				)
			},
			args: args{
				id:  1,
				all: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "User with no credentials",
			setup: func() {
				mocker := mockPool.EXPECT()
				mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByIdSql), int64(1)).Return(
					pgxpoolmock.NewRow(int64(1), "test-id", []byte("test-id"), "name", "display"),
				)
				mocker.Query(gomock.Any(), pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_credentials.*)"), pgtype.Int8{Int64: 1, Valid: true}).Return(
					pgxpoolmock.NewRows(credentialRows).AddRow(
						[]byte{}, int64(0), int32(0), []byte{}, "", []byte{}, []byte{}, []byte{}, []byte{}, []byte{},
					).ToPgxRows(),
					pgx.ErrNoRows,
				)
			},
			args: args{
				id:  1,
				all: false,
			},
			want:    buildUserModel(1, "test-id", "name", "display"),
			wantErr: false,
		},
		{
			name: "User with all credentials",
			setup: func() {
				mocker := mockPool.EXPECT()
				mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByIdSql), int64(1)).Return(
					pgxpoolmock.NewRow(int64(1), "test-id", []byte("test-id"), "name", "display"),
				)
				// Ensure that the json meta->>'status' = true condition is not present
				mocker.Query(gomock.Any(), pgxpoolmock.QueryContains(`(?ms:SELECT.*FROM webauthn_credentials.*WHERE user_id = \$1\s+ORDER BY)`), pgtype.Int8{Int64: 1, Valid: true}).Return(
					pgxpoolmock.NewRows(credentialRows).AddRow(
						mockCredentialRow("c1", true, "c1-nickname"),
					).AddRow(
						mockCredentialRow("c2", false, "c2-nickname"),
					).ToPgxRows(),
					nil,
				)
			},
			args: args{
				id:  1,
				all: true,
			},
			want: buildUserModel(
				1,
				"test-id",
				"name",
				"display",
				buildCredentialModel("c1", true, "c1-nickname"),
				buildCredentialModel("c2", false, "c2-nickname"),
			),
			wantErr: false,
		},
		{
			name: "User with only active credentials",
			setup: func() {
				mocker := mockPool.EXPECT()
				mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByIdSql), int64(1)).Return(
					pgxpoolmock.NewRow(int64(1), "test-id", []byte("test-id"), "name", "display"),
				)
				// Ensure that the json meta->>'status' = 'active' condition is present
				mocker.Query(gomock.Any(), pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_credentials.*meta->>'status' = 'active'.*)"), pgtype.Int8{Int64: 1, Valid: true}).Return(
					pgxpoolmock.NewRows(credentialRows).AddRow(
						mockCredentialRow("c1", true, "c1-nickname"),
					).AddRow(
						mockCredentialRow("c2", true, "c2-nickname"),
					).ToPgxRows(),
					nil,
				)
			},
			args: args{
				id:  1,
				all: false,
			},
			want: buildUserModel(
				1,
				"test-id",
				"name",
				"display",
				buildCredentialModel("c1", true, "c1-nickname"),
				buildCredentialModel("c2", true, "c2-nickname"),
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			tt.setup()

			s, err := New(context.Background())
			if err != nil {
				t.Errorf("New() error = %v, want nil", err)
			}

			got, err := s.GetUserWithCredentialsByID(tt.args.id, tt.args.all)
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
	type setup func()
	type args struct {
		ref string
		all bool
	}
	tests := []struct {
		name    string
		setup   setup
		args    args
		want    *UserModel
		wantErr bool
	}{
		{
			name: "User not found",
			setup: func() {
				mockPool.EXPECT().QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByRefSql), "ref").Return(
					pgxpoolmock.NewRow(int64(0), "", []byte{}, "", "").WithError(pgx.ErrNoRows),
				)
			},
			args: args{
				ref: "ref",
				all: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "User with no credentials",
			setup: func() {
				mocker := mockPool.EXPECT()
				mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByRefSql), "test-id").Return(
					pgxpoolmock.NewRow(int64(1), "test-id", []byte("test-id"), "name", "display"),
				)
				mocker.Query(gomock.Any(), pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_credentials.*)"), pgtype.Int8{Int64: 1, Valid: true}).Return(
					pgxpoolmock.NewRows(credentialRows).AddRow(
						[]byte{}, int64(0), int32(0), []byte{}, "", []byte{}, []byte{}, []byte{}, []byte{}, []byte{},
					).ToPgxRows(),
					pgx.ErrNoRows,
				)
			},
			args: args{
				ref: "test-id",
				all: false,
			},
			want:    buildUserModel(1, "test-id", "name", "display"),
			wantErr: false,
		},
		{
			name: "User with all credentials",
			setup: func() {
				mocker := mockPool.EXPECT()
				mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByRefSql), "test-id").Return(
					pgxpoolmock.NewRow(int64(1), "test-id", []byte("test-id"), "name", "display"),
				)
				// Ensure that the json meta->>'status' = true condition is not present
				mocker.Query(gomock.Any(), pgxpoolmock.QueryContains(`(?ms:SELECT.*FROM webauthn_credentials.*WHERE user_id = \$1\s+ORDER BY)`), pgtype.Int8{Int64: 1, Valid: true}).Return(
					pgxpoolmock.NewRows(credentialRows).AddRow(
						mockCredentialRow("c1", true, "c1-nickname"),
					).AddRow(
						mockCredentialRow("c2", false, "c2-nickname"),
					).ToPgxRows(),
					nil,
				)
			},
			args: args{
				ref: "test-id",
				all: true,
			},
			want: buildUserModel(
				1,
				"test-id",
				"name",
				"display",
				buildCredentialModel("c1", true, "c1-nickname"),
				buildCredentialModel("c2", false, "c2-nickname"),
			),
			wantErr: false,
		},
		{
			name: "User with only active credentials",
			setup: func() {
				mocker := mockPool.EXPECT()
				mocker.QueryRow(gomock.Any(), pgxpoolmock.QueryContains(getUserByRefSql), "test-id").Return(
					pgxpoolmock.NewRow(int64(1), "test-id", []byte("test-id"), "name", "display"),
				)
				// Ensure that the json meta->>'status' = 'active' condition is present
				mocker.Query(gomock.Any(), pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_credentials.*meta->>'status' = 'active'.*)"), pgtype.Int8{Int64: 1, Valid: true}).Return(
					pgxpoolmock.NewRows(credentialRows).AddRow(
						mockCredentialRow("c1", true, "c1-nickname"),
					).AddRow(
						mockCredentialRow("c2", true, "c2-nickname"),
					).ToPgxRows(),
					nil,
				)
			},
			args: args{
				ref: "test-id",
				all: false,
			},
			want: buildUserModel(
				1,
				"test-id",
				"name",
				"display",
				buildCredentialModel("c1", true, "c1-nickname"),
				buildCredentialModel("c2", true, "c2-nickname"),
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			tt.setup()

			s, err := New(context.Background())
			if err != nil {
				t.Errorf("New() error = %v, want nil", err)
			}

			got, err := s.GetUserWithCredentialsByRef(tt.args.ref, tt.args.all)
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
	type setup func()
	type args struct {
		user       *UserModel
		credential *webauthn.Credential
	}
	tests := []struct {
		name    string
		setup   setup
		args    args
		wantErr bool
	}{
		{
			name: "Insert credential success",
			setup: func() {
				mockPool.EXPECT().QueryRow(
					gomock.Any(),
					pgxpoolmock.QueryContains("(?ms:INSERT INTO webauthn_credentials.*)"),
					[]byte("credential-id"),
					gomock.Any(),
					[]byte("public-key"),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(pgxpoolmock.NewRow([]byte("credential-id"), pgtype.Int8{Int64: 1, Valid: true}, int32(0), []byte{}, pgtype.Text{String: "none", Valid: true}, []byte{}, []byte{}, []byte{}, []byte{}, []byte{}))
			},
			args: args{
				user:       buildUserModel(1, "test-id", "name", "display"),
				credential: buildWebAuthnCredential("credential-id"),
			},
			wantErr: false,
		},
		{
			name:  "Insert credential invalid user",
			setup: func() {},
			args: args{
				user:       buildUserModel(0, "test-id", "name", "display"),
				credential: buildWebAuthnCredential("credential-id"),
			},
			wantErr: true,
		},
		{
			name: "Insert credential error",
			setup: func() {
				mockPool.EXPECT().QueryRow(
					gomock.Any(),
					pgxpoolmock.QueryContains("(?ms:INSERT INTO webauthn_credentials.*)"),
					[]byte("credential-id"),
					gomock.Any(),
					[]byte("public-key"),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(
					pgxpoolmock.NewRow([]byte{}, pgtype.Int8{Int64: 0, Valid: false}, int32(0), []byte{}, pgtype.Text{String: "", Valid: false}, []byte{}, []byte{}, []byte{}, []byte{}, []byte{}).WithError(fmt.Errorf("query failed")),
				)
			},
			args: args{
				user:       buildUserModel(1, "test-id", "name", "display"),
				credential: buildWebAuthnCredential("credential-id"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			tt.setup()

			s, err := New(context.Background())
			if err != nil {
				t.Errorf("New() error = %v, want nil", err)
			}

			if err := s.InsertCredential(tt.args.user, tt.args.credential); (err != nil) != tt.wantErr {
				t.Errorf("CredentialService.InsertCredential() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCredentialService_IncrementCredentialUseCounter(t *testing.T) {
	type setup func()
	type args struct {
		credentialID []byte
	}
	tests := []struct {
		name    string
		setup   setup
		args    args
		want    int32
		wantErr bool
	}{
		{
			name: "Increment use counter success",
			setup: func() {
				mockPool.EXPECT().QueryRow(gomock.Any(), pgxpoolmock.QueryContains("(?ms:UPDATE webauthn_credentials.*)"), []byte("credential-id")).Return(
					pgxpoolmock.NewRow(int32(1)),
				)
			},
			args: args{
				credentialID: []byte("credential-id"),
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Increment use counter error",
			setup: func() {
				mockPool.EXPECT().QueryRow(gomock.Any(), pgxpoolmock.QueryContains("(?ms:UPDATE webauthn_credentials.*)"), []byte("credential-id")).Return(
					pgxpoolmock.NewRow(int32(0)).WithError(fmt.Errorf("query failed")),
				)
			},
			args: args{
				credentialID: []byte("credential-id"),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			tt.setup()

			s, err := New(context.Background())
			if err != nil {
				t.Errorf("New() error = %v, want nil", err)
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
