package credential_service

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"blacksmithlabs.dev/webauthn-k8s/auth/database"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
	"blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/milqa/pgxpoolmock"
)

var mockPool *pgxpoolmock.MockPgxIface

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
	mocker.QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_users.*)"),
		"123",
	).Return(
		pgxpoolmock.NewRow(int64(0), "", []byte{}, "", "").WithError(pgx.ErrNoRows),
	)
	mocker.QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:INSERT INTO webauthn_users.*)"),
		"123",
		[]byte("123"),
		"User Name",
		"Display Name",
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

func TestCredentialService_UpsertUser_UserFound(t *testing.T) {
	// Given
	setupTest(t)

	mocker := mockPool.EXPECT()
	mocker.Begin(gomock.Any()).Return(mockPool, nil)
	mocker.Commit(gomock.Any()).Return(nil)
	mocker.Rollback(gomock.Any()).Return(nil)
	mocker.QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_users.*)"),
		"123",
	).Return(
		pgxpoolmock.NewRow(int64(1), "123", []byte("123"), "User Name", "Display Name"),
	)
	// Insert query should not be called

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

func TestCredentialService_UpsertUser_Error(t *testing.T) {
	// Given
	setupTest(t)

	mocker := mockPool.EXPECT()
	mocker.Begin(gomock.Any()).Return(mockPool, nil)
	// Commit should not be called
	mocker.Rollback(gomock.Any()).Return(nil)
	mocker.QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:SELECT.*FROM webauthn_users.*)"),
		"123",
	).Return(
		pgxpoolmock.NewRow(int64(0), "", []byte{}, "", "").WithError(pgx.ErrNoRows),
	)
	mocker.QueryRow(
		gomock.Any(),
		pgxpoolmock.QueryContains("(?ms:INSERT INTO webauthn_users.*)"),
		"123",
		[]byte("123"),
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

func TestCredentialService_GetUserWithCredentialsByID(t *testing.T) {
	type fields struct {
		ctx     context.Context
		conn    database.DBConn
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
		conn    database.DBConn
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
		conn    database.DBConn
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
		conn    database.DBConn
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
