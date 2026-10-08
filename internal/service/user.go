package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/encourage13/todo-api/internal/model"
)

type userRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type UserService struct {
	store userRepository
}

func NewUserService(repo userRepository) *UserService {
	return &UserService{store: repo}
}

func (u *UserService) Register(ctx context.Context, email, password string) (*model.User, error) {
	e, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	if !(len(password) >= 8 && len(password) <= 72) {
		return nil, fmt.Errorf("%w: password must be 8-72 characters long", model.ErrInvalidInput)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("couldnt generate hash: %w", err)
	}

	user := model.User{Email: e, PasswordHash: string(hash)}
	err = u.store.Create(ctx, &user)

	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			return nil, model.ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func (u *UserService) Login(ctx context.Context, email, password string) (*model.User, error) {
	if len(password) > 72 {
		return nil, model.ErrInvalidCredentials
	}

	e, err := normalizeEmail(email)
	if err != nil {
		return nil, model.ErrInvalidCredentials
	}

	user, err := u.store.GetByEmail(ctx, e)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, model.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("compare pass error: %w", err)
	}
	return user, nil
}

func normalizeEmail(email string) (string, error) {
	e := strings.TrimSpace(email)
	e = strings.ToLower(e)

	addr, err := mail.ParseAddress(e)
	if err != nil {
		return "", fmt.Errorf("%w: invalid email", model.ErrInvalidInput)
	}
	if e != addr.Address {
		return "", fmt.Errorf("%w: invalid email", model.ErrInvalidInput)
	}
	return e, nil
}
