package model

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(id int64, name, email, password string, now time.Time) *User {
	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
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
