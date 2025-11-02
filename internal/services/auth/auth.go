package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, id int64) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWrongApp           = errors.New("wrong application")
)

// New creates a new Auth service instance.
//
// Parameters:
// - log:  A pointer to a slog.Logger for logging purposes.
// - userSaver: An implementation of the UserSaver interface for saving user data.
// - userProvider: An implementation of the UserProvider interface for retrieving user data.
// - appProvider: An implementation of the AppProvider interface for retrieving application data.
// - tokenTTL: A time.Duration representing the time-to-live for authentication tokens.
//
// Returns:
// - A pointer to an Auth service instance.
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {

		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid password", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, int64(appID))
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		log.Error("failed to get app", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully", slog.Int64("user_id", user.ID), slog.Int("app_id", app.ID))

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil

}

// RegisterNewUser registers a new user in the system.
//
// Parameters:
// - ctx: The context for the operation.
// - email: The email address of the new user.
// - pass: The password for the new user.
//
// Returns:
// - userID: An integer representing the unique ID of the newly registered user.
// - err: An error object if the registration fails, otherwise nil.
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	pass string,
) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")
	return id, nil
}

// IsAdmin checks if a user has administrative privileges.
//
// Parameters:
// - ctx: The context for the operation.
// - userID: The unique ID of the user to be checked.
//
// Returns:
// - isAdmin: A boolean indicating whether the user is an admin (true) or not (false).
// - err: An error object if the check fails, otherwise nil.
func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")
	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
