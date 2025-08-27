package data

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"simplewebapi.moviedb/internal/validator"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type password struct {
	plaintext *string // bắt data đổ về từ sql nil  trực quan hơn ""
	hash      []byte
}

func (p *password) Set(raw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), 11)
	if err != nil {
		return err
	}

	p.plaintext = &raw
	p.hash = hashed
	return nil
}
func (p *password) Match(s string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(s))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Match(email, validator.EmailRX), "email", "invalid email address")
}
func ValidatePassword(v *validator.Validator, pw string) {
	v.Check(pw != "", "password", "must be provided")
	v.Check(len(pw) >= 6, "password", "must be at least 6 characters")
	v.Check(len(pw) <= 72, "password", "must not be more than 72 characters")
}
func ValidateUser(v *validator.Validator, u *User) {
	v.Check(u.Name != "", "name", "must be provided")
	v.Check(len(u.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, u.Email)

	if u.Password.plaintext != nil {
		ValidatePassword(v, *u.Password.plaintext)
	}
	if u.Password.hash == nil {
		panic("missing password hash for user")
	}

}
