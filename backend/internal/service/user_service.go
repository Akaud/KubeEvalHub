package service

import (
	"context"
	"errors"
	"strings"

	"backend/internal/model"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrInvalidUserName     = errors.New("invalid user name")
	ErrInvalidUserEmail    = errors.New("invalid user email")
	ErrInvalidUserPassword = errors.New("invalid user password")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already exists")
)

type UserRepository interface {
	Create(ctx context.Context, name, email, password string) (*model.User, error)
	Update(ctx context.Context, id int64, name, email, password string) (*model.User, error)
	Delete(ctx context.Context, id int64) error
	Patch(ctx context.Context, id int64, name, email, password *string) (*model.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrInvalidUserName
	}
	return name, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", ErrInvalidUserEmail
	}
	return email, nil
}

func normalizePassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", ErrInvalidUserPassword
	}
	return password, nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*model.User, error) {
	name, err := normalizeName(name)
	if err != nil {
		return nil, err
	}

	email, err = normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	password, err = normalizePassword(password)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Create(ctx, name, email, hashedPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmailAlreadyExists):
			return nil, ErrEmailAlreadyExists
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, name, email, password string) (*model.User, error) {
	if id <= 0 {
		return nil, ErrInvalidUserID
	}

	name, err := normalizeName(name)
	if err != nil {
		return nil, err
	}

	email, err = normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	password, err = normalizePassword(password)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Update(ctx, id, name, email, hashedPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		case errors.Is(err, repository.ErrEmailAlreadyExists):
			return nil, ErrEmailAlreadyExists
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidUserID
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return ErrUserNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *UserService) PatchUser(ctx context.Context, id int64, name, email, password *string) (*model.User, error) {
	if id <= 0 {
		return nil, ErrInvalidUserID
	}

	var normalizedName *string
	if name != nil {
		v, err := normalizeName(*name)
		if err != nil {
			return nil, err
		}
		normalizedName = &v
	}

	var normalizedEmail *string
	if email != nil {
		v, err := normalizeEmail(*email)
		if err != nil {
			return nil, err
		}
		normalizedEmail = &v
	}

	var normalizedPassword *string
	if password != nil {
		v, err := normalizePassword(*password)
		if err != nil {
			return nil, err
		}

		hashed, err := hashPassword(v)
		if err != nil {
			return nil, err
		}
		normalizedPassword = &hashed
	}

	user, err := s.repo.Patch(ctx, id, normalizedName, normalizedEmail, normalizedPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		case errors.Is(err, repository.ErrEmailAlreadyExists):
			return nil, ErrEmailAlreadyExists
		default:
			return nil, err
		}
	}

	return user, nil
}
