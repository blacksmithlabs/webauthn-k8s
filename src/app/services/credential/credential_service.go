package credential_service

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/milqa/pgxpoolmock/sqlc"

	"blacksmithlabs.dev/webauthn-k8s/app/database"
	"blacksmithlabs.dev/webauthn-k8s/app/utils"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
	"blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
)

// CredentialService provides methods for interacting with user credentials
type CredentialService struct {
	ctx     context.Context
	conn    sqlc.DBTX
	queries *credentials.Queries
}

var getDbConn func(context.Context) (sqlc.DBTX, error) = func(ctx context.Context) (sqlc.DBTX, error) {
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

// Parse a list of user credentials from the database into a user model
func parseUserCredentialList(userCredentials []credentials.ResultUserCredentialRow) (*UserModel, error) {
	var user *UserModel
	for _, row := range userCredentials {
		if user == nil {
			user = UserModelFromDatabase(*row.GetUser())
		}
		credential, err := CredentialModelFromDatabase(*row.GetCredential())
		if err != nil {
			return nil, err
		}
		user.addCredential(credential.Credential)
	}

	return user, nil
}

// UpsertUser creates or updates a user in the database based on the provided user information from the DTO
func (s *CredentialService) UpsertUser(userDto dto.RegistrationUserInfo) (*UserModel, error) {
	// TODO generate a random byte array for the raw ID

	user, err := s.queries.UpsertUser(s.ctx, credentials.UpsertUserParams{
		RefID:       userDto.UserID,
		RawID:       []byte(userDto.UserID),
		Name:        userDto.UserName,
		DisplayName: userDto.DisplayName,
	})
	if err != nil {
		return nil, err
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

// GetUserWithCredentialsByID retrieves a user from the database based on the provided ID and includes the user's credentials
func (s *CredentialService) GetUserWithCredentialsByID(id int64) (*UserModel, error) {
	userCredentials, err := s.queries.GetUserWithCredentialsByID(s.ctx, id)
	if err != nil {
		return nil, err
	}

	if userCredentials == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Type casting is stupid. I don't know why it won't work with an array and this is necessary
	return parseUserCredentialList(utils.Map(userCredentials, func(row credentials.GetUserWithCredentialsByIDRow) credentials.ResultUserCredentialRow {
		return row
	}))
}

// GetUserWithCredentialsByRef retrieves a user from the database based on the provided reference and includes the user's credentials
func (s *CredentialService) GetUserWithCredentialsByRef(ref string) (*UserModel, error) {
	userCredentials, err := s.queries.GetUserWithCredentialsByRef(s.ctx, ref)
	if err != nil {
		return nil, err
	}

	if userCredentials == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Type casting is stupid. I don't know why it won't work with an array and this is necessary
	return parseUserCredentialList(utils.Map(userCredentials, func(row credentials.GetUserWithCredentialsByRefRow) credentials.ResultUserCredentialRow {
		return row
	}))
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
