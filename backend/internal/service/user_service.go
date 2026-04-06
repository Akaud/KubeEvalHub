package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUserID          = errors.New("invalid user id")
	ErrInvalidUserName        = errors.New("invalid user name")
	ErrInvalidUserEmail       = errors.New("invalid user email")
	ErrInvalidUserPassword    = errors.New("invalid user password")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidLoginIdentifier = errors.New("invalid login identifier")
	ErrInvalidToken           = errors.New("invalid token")
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists     = errors.New("email already exists")
)

type UserRepository interface {
	Create(ctx context.Context, name, email, password string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByName(ctx context.Context, name string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, name, email, password string) (*model.User, error)
	Delete(ctx context.Context, id int64) error
	Patch(ctx context.Context, id int64, name, email, password *string) (*model.User, error)
}

type UserService struct {
	repo      UserRepository
	jwtSecret []byte
}

func NewUserService(repo UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
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

func normalizeLoginIdentifier(identifier string) (string, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return "", ErrInvalidLoginIdentifier
	}
	return identifier, nil
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

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	if id <= 0 {
		return nil, ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
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

type authClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *UserService) AuthenticateUser(ctx context.Context, identifier, password string) (string, error) {
	identifier, err := normalizeLoginIdentifier(identifier)
	if err != nil {
		return "", err
	}

	password, err = normalizePassword(password)
	if err != nil {
		return "", err
	}

	var user *model.User

	if strings.Contains(identifier, "@") {
		email, err := normalizeEmail(identifier)
		if err != nil {
			return "", ErrInvalidCredentials
		}

		user, err = s.repo.GetByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return "", ErrInvalidCredentials
			}
			return "", err
		}
	} else {
		name, err := normalizeName(identifier)
		if err != nil {
			return "", ErrInvalidCredentials
		}

		user, err = s.repo.GetByName(ctx, name)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return "", ErrInvalidCredentials
			}
			return "", err
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	now := time.Now().UTC()
	expiresAt := now.Add(20 * time.Minute)

	claims := authClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *UserService) ValidateToken(ctx context.Context, tokenString string) (int64, error) {
	if strings.TrimSpace(tokenString) == "" {
		return 0, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &authClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*authClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	if claims.UserID <= 0 {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}
