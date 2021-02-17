package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	BlogPosts Posts     `json:"posts" has_many:"posts"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Posts - Return a collection of posts belong to user
func (u *User) Posts(tx *pop.Connection) (*Posts, error) {
	posts := &Posts{}
	err := tx.Eager().All(posts)

	return posts, err
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.EmailIsPresent{Field: u.Email},
		&validators.StringLengthInRange{Field: u.Password},
	), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create - create user with hashed password
func (u *User) Create(tx *pop.Connection) (*validate.Errors, error) {
	u.Email = strings.ToLower(u.Email)
	// check email is exist
	verrs := validate.NewErrors()

	hashedPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

	// if the bcrypt has any error
	if bcryptErr != nil {
		verrs.Add("password", "There is a problem when performing the password hashing.")

		return verrs, errors.WithStack(bcryptErr)
	}
	u.Password = string(hashedPassword)

	// create user

	return verrs, tx.Create(u)
}
