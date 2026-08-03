package users

import (
	"errors"
)

type UserRepository interface {
	Save(user User) (User, error)
	FindByEmail(email string) (User, error)
}

var ErrEmailTaken = errors.New("email already registered")

// Creates a new User account, and stores it to the repository.
func NewUser(
	userRepository UserRepository,
	email string,
	displayName string,
	plaintextPassword string,
) (User, error) {
	// Check that no user with the given email exists in the repo yet
	_, err := userRepository.FindByEmail(email)
	if err == nil { // No error means that the email has already been taken
		return User{}, ErrEmailTaken
	} else if !errors.Is(err, ErrNoSuchUser) { // Return any *unexpected* errors
		return User{}, err
	}

	passwordHash, err := HashPassword(plaintextPassword)
	if err != nil {
		return User{}, err
	}

	// Begin to create the User record, filling the given fields, and letting
	// the repository populate the rest
	var user = User{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
	}

	return userRepository.Save(user)
}
