package model

import "time"

type User struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	Password          string     `json:"-"`
	EmailVerified     bool       `json:"email_verified"`
	EmailVerifyToken  string     `json:"-"`
	VerifyTokenExpiry *time.Time `json:"-"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func NewUser(id int64, name, email, password string, now time.Time) *User {
	return &User{
		ID:            id,
		Name:          name,
		Email:         email,
		Password:      password,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func NewUserWithVerification(id int64, name, email, password, verifyToken string, tokenExpiry time.Time, now time.Time) *User {
	return &User{
		ID:                id,
		Name:              name,
		Email:             email,
		Password:          password,
		EmailVerified:     false,
		EmailVerifyToken:  verifyToken,
		VerifyTokenExpiry: &tokenExpiry,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (u *User) Update(name, email, password string, now time.Time) {
	u.Name = name
	u.Email = email
	u.Password = password
	u.UpdatedAt = now
}

func (u *User) Patch(name, email, password *string, now time.Time) bool {
	changed := false

	if name != nil && u.Name != *name {
		u.Name = *name
		changed = true
	}

	if email != nil && u.Email != *email {
		u.Email = *email
		changed = true
	}

	if password != nil && u.Password != *password {
		u.Password = *password
		changed = true
	}

	if changed {
		u.UpdatedAt = now
	}

	return changed
}

func (u *User) MarkEmailVerified(now time.Time) {
	u.EmailVerified = true
	u.EmailVerifyToken = ""
	u.VerifyTokenExpiry = nil
	u.UpdatedAt = now
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerified
}

func (u *User) HasValidVerificationToken() bool {
	if u.EmailVerified {
		return false
	}
	if u.EmailVerifyToken == "" {
		return false
	}
	if u.VerifyTokenExpiry == nil {
		return false
	}
	return time.Now().UTC().Before(*u.VerifyTokenExpiry)
}

func (u *User) SetVerificationToken(token string, expiry time.Time, now time.Time) {
	u.EmailVerifyToken = token
	u.VerifyTokenExpiry = &expiry
	u.EmailVerified = false
	u.UpdatedAt = now
}

func (u *User) ClearVerificationToken(now time.Time) {
	u.EmailVerifyToken = ""
	u.VerifyTokenExpiry = nil
	u.UpdatedAt = now
}
