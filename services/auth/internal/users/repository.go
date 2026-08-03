package users

import (
	"errors"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// All possible errors that can be returned by the repository
var ErrNoSuchUser = errors.New("No such user exists within the database")
var ErrInvalidDisplayNameLength = errors.New("len(display_name) must fall between 1 and 50 inclusive.")
var ErrEmptyPassword = errors.New("password_hash must not be empty.")
var ErrInvalidEmailLength = errors.New("An email must not be empty, and must contain no more than 320 characters.")
var ErrInvalidEmailFormat = errors.New("Incorrect email format")

// In-memory implementation of a User repository for testing and development purposes only.
type InMemoryUserRepository struct {
	userIdMap    map[uuid.UUID]User
	userEmailMap map[string]User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		userIdMap:    make(map[uuid.UUID]User),
		userEmailMap: make(map[string]User),
	}
}

// Base methods

// Persists a User to the repository
func (repo *InMemoryUserRepository) Save(user User) (User, error) {
	if err := validateMandatoryAttributes(user); err != nil {
		return User{}, err
	}
	user = populateOptionalAttributes(user)

	repo.userIdMap[user.Id] = user
	repo.userEmailMap[user.Email] = user

	return user, nil
}

// Finds and retrieves a User by their email
func (repo *InMemoryUserRepository) FindByEmail(email string) (User, error) {
	if err := validateEmailFormat(email); err != nil {
		return User{}, err
	}

	var user = repo.userEmailMap[email]
	if user == (User{}) {
		return user, ErrNoSuchUser
	}
	return user, nil
}

// Finds and retrieves a User by their ID
func (repo *InMemoryUserRepository) FindById(id uuid.UUID) (User, error) {
	var user = repo.userIdMap[id]
	if user == (User{}) {
		return user, ErrNoSuchUser
	}
	return user, nil
}

// Retrieves all User records from the repository
func (repo *InMemoryUserRepository) FindAll() []User {
	var userSlice = slices.Collect(maps.Values(repo.userIdMap))
	if userSlice == nil { // Convert a nil pointer into an empty slice
		userSlice = []User{}
	}
	return userSlice
}

// Deletes the record of the User with the given email.
func (repo *InMemoryUserRepository) DeleteByEmail(email string) error {
	// Validate that the user exists lazily by using findByEmail()
	user, err := repo.FindByEmail(email)
	if err != nil {
		return err
	}
	repo.userEmailMap[user.Email] = User{}
	repo.userIdMap[user.Id] = User{}
	return nil
}

// Deletes the record of the User with the given ID.
func (repo *InMemoryUserRepository) DeleteById(id uuid.UUID) error {
	// Validate that the user exists lazily by using findById()
	user, err := repo.FindById(id)
	if err != nil {
		return err
	}
	repo.userEmailMap[user.Email] = User{}
	repo.userIdMap[user.Id] = User{}
	return nil
}

// Helper methods

// Helper method which ensures that all caller-specified attributes of a User
// are in the correct format.
func validateMandatoryAttributes(user User) error {
	if user.DisplayName == "" || len(user.DisplayName) > 50 {
		return ErrInvalidDisplayNameLength
	}
	if user.PasswordHash == "" {
		return ErrEmptyPassword
	}
	if err := validateEmailFormat(user.Email); err != nil {
		return err
	}

	return nil
}

// Helper method dedicated to checking the format of an email address
func validateEmailFormat(email string) error {
	// An email must not be empty, but must contain at most 320 chars.
	if email == "" || len(email) > 320 {
		return ErrInvalidEmailLength
	}

	// Split used for @ check, as only one @ can be present
	var emailComponents = strings.Split(email, "@")
	if len(emailComponents) != 2 {
		return ErrInvalidEmailFormat
	}

	// Cut used for . check, as any number of .s could also be present in topLevelDomain
	var localPart = emailComponents[0]
	var domain, topLevelDomain, _ = strings.Cut(emailComponents[1], ".")

	if localPart == "" || domain == "" || topLevelDomain == "" {
		return ErrInvalidEmailFormat
	}

	return nil

}

// Helper method used to populate the attributes which don't need to be
// specified by the caller.
func populateOptionalAttributes(user User) User {
	if user.Id == (uuid.UUID{}) {
		user.Id = uuid.New()
	}

	// Note: updated_at is always updated, regardless of its state
	if time.Time.Equal(user.CreatedAt, time.Time{}) {
		user.CreatedAt = time.Now()
		user.UpdatedAt = user.CreatedAt
	} else {
		user.UpdatedAt = time.Now()
	}

	return user
}
