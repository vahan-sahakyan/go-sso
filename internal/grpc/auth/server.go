package auth

import (
	"context"

	ssov1 "github.com/vahan-sahakyan/go-protobufs/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	EMPTY_VALUE = 0
)

type Auth interface {
	// Login authenticates a user and returns a token if successful.
	//
	// Parameters:
	// - email: The user's email address.
	// - password: The user's password.
	// - appID: The application ID for which the token is requested.
	//
	// Returns:
	// - token: A string representing the authentication token.
	// - err: An error object if the login fails, otherwise nil.
	Login(
		email string,
		password string,
		appID int,
	) (token string, err error)
	// RegisterNewUser registers a new user in the system.
	//
	// Parameters:
	// - email: The email address of the new user.
	// - password: The password for the new user.
	//
	// Returns:
	// - userID: An integer representing the unique ID of the newly registered user.
	// - err: An error object if the registration fails, otherwise nil.
	RegisterNewUser(
		email string,
		password string,
	) (userID int64, err error)
	// IsAdmin checks if a user has administrative privileges.
	//
	// Parameters:
	// - userID: The unique ID of the user to be checked.
	//
	// Returns:
	// - isAdmin: A boolean indicating whether the user is an admin (true) or not (false).
	// - err: An error object if the check fails, otherwise nil.
	IsAdmin(
		ctx context.Context,
		userID int64,
	) (bool, error)
}

type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth})
}

func (s *serverApi) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(req.GetEmail(), req.GetPassword(), int(req.GetAppId()))

	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(req.GetEmail(), req.GetPassword())
	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == EMPTY_VALUE {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == EMPTY_VALUE {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	return nil
}
