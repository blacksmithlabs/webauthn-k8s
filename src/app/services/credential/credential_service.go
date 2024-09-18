package credential_service

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"blacksmithlabs.dev/webauthn-k8s/app/database"
	"blacksmithlabs.dev/webauthn-k8s/app/utils"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
	"blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
)

// CredentialService provides methods for interacting with user credentials
type CredentialService struct {
	ctx     context.Context
	conn    database.DBConn
	queries *credentials.Queries
}

var getDbConn func(context.Context) (database.DBConn, error) = func(ctx context.Context) (database.DBConn, error) {
	return database.ConnectDb(ctx)
}

// New creates a new CredentialService instance
func New(ctx context.Context) (*CredentialService, error) {
	pool, err := getDbConn(ctx)
	if err != nil {
		return nil, err
	}

	queries := credentials.New(pool)

	return &CredentialService{
		ctx:     ctx,
		conn:    pool,
		queries: queries,
	}, nil
}

// UpsertUser creates or updates a user in the database based on the provided user information from the DTO
func (s *CredentialService) UpsertUser(userDto dto.RegistrationUserInfo) (*UserModel, error) {
	// TODO generate a random byte array for the raw ID

	tx, err := s.conn.Begin(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	txn := s.queries.WithTx(tx)

	defer tx.Rollback(s.ctx)

	user, err := txn.GetUserByRef(s.ctx, userDto.UserID)
	if err == pgx.ErrNoRows {
		// User does not exist, create a new user
		user, err = txn.InsertUser(s.ctx, credentials.InsertUserParams{
			RefID:       userDto.UserID,
			RawID:       []byte(userDto.UserID),
			Name:        userDto.UserName,
			DisplayName: userDto.DisplayName,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	err = tx.Commit(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return UserModelFromDatabase(user), nil
}

// GetUserByID retrieves a user from the database based on the provided ID
func (s *CredentialService) GetUserByID(id int64) (*UserModel, error) {
	user, err := s.queries.GetUserByID(s.ctx, id)
	if err != nil {
		return nil, err
	}

	return UserModelFromDatabase(user), nil
}

// GetUserByRef retrieves a user from the database based on the provided reference
func (s *CredentialService) GetUserByRef(ref string) (*UserModel, error) {
	user, err := s.queries.GetUserByRef(s.ctx, ref)
	if err != nil {
		return nil, err
	}

	return UserModelFromDatabase(user), nil
}

func (s *CredentialService) addUserCredentialList(user *UserModel) error {
	userCredentials, err := s.queries.ListCredentialsByUser(s.ctx, pgtype.Int8{
		Int64: user.ID,
		Valid: true,
	})
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	for _, row := range userCredentials {
		credential, err := CredentialModelFromDatabase(row)
		if err != nil {
			return fmt.Errorf("failed to parse credential: %w", err)
		}

		user.addCredential(credential.Credential)
	}

	return nil
}

// GetUserWithCredentialsByID retrieves a user from the database based on the provided ID and includes the user's credentials
func (s *CredentialService) GetUserWithCredentialsByID(id int64) (*UserModel, error) {
	user, err := s.queries.GetUserByID(s.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	userModel := UserModelFromDatabase(user)
	err = s.addUserCredentialList(userModel)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return userModel, nil
}

// GetUserWithCredentialsByRef retrieves a user from the database based on the provided reference and includes the user's credentials
func (s *CredentialService) GetUserWithCredentialsByRef(ref string) (*UserModel, error) {
	user, err := s.queries.GetUserByRef(s.ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	userModel := UserModelFromDatabase(user)
	err = s.addUserCredentialList(userModel)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return userModel, nil
}

// InsertCredential inserts a credential into the database for the provided user
func (s *CredentialService) InsertCredential(user *UserModel, credential *webauthn.Credential) error {
	model := &CredentialModel{
		Credential: *credential,
		User:       utils.Relationship[UserModel]{Loaded: true, Value: *user},
	}
	params, err := model.ToInsertParams()
	if err != nil {
		return fmt.Errorf("failed to convert credential to model: %w", err)
	}

	if _, err := s.queries.InsertCredential(s.ctx, *params); err != nil {
		return fmt.Errorf("data access error: %w", err)
	}

	return nil
}

// IncrementCredentialUseCounter increments the use counter for a credential in the database
func (s *CredentialService) IncrementCredentialUseCounter(credentialID []byte) (int32, error) {
	useCount, err := s.queries.IncrementCredentialUseCounter(s.ctx, credentialID)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	return useCount, nil
}
