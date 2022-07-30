package core

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"

	jwt "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/pbkdf2"
)

const (
	passwordSaltLength = 12

	defaultHashKeyLength = 12
	defaultHashRounds    = 4096

	audAPIAdminV1 = "api/admin/v1"
)

var (
	tenYears = 10 * 365 * 24 * time.Hour
)

type AuthService struct {
	jwtIssuer          string
	jwtSecret          []byte
	jwtDurationAccess  time.Duration
	jwtDurationRefresh time.Duration
	passwordSalt       []byte

	refreshTokenStore RefreshTokenStore
	passwordStore     PasswordStore
	userStore         UserStore

	emailClient interface{}
}

func NewAuthService(config common.Config, userStore UserStore, passwordStore PasswordStore, refreshTokenStore RefreshTokenStore) AuthService {
	return AuthService{
		jwtIssuer:          config.Server.BaseURL,
		jwtSecret:          []byte(config.Auth.JWTSecret),
		jwtDurationAccess:  config.Auth.AccessTokenExpiry(),
		jwtDurationRefresh: config.Auth.RefreshTokenExpiry(),
		passwordSalt:       []byte(config.Auth.PasswordSalt),

		refreshTokenStore: refreshTokenStore,
		passwordStore:     passwordStore,
		userStore:         userStore,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, reg auth.Registration) (auth.User, error) {
	user := auth.User{
		Name:  reg.Username,
		Email: reg.Email,
	}
	if err := user.Validate(); err != nil {
		return auth.User{}, common.WrapErr(fmt.Errorf("failed to make user: %s", err), common.ErrCodeBadRequest)
	}

	randomSalt := make([]byte, passwordSaltLength)
	if _, err := rand.Read(randomSalt); err != nil {
		return auth.User{}, common.WrapErr(fmt.Errorf("cannot make password: %s", err), common.ErrCodeServer)
	}

	password, err := s.makePassword(randomSalt, reg.Credentials, passwordOptions{})
	if err != nil {
		return auth.User{}, err
	}

	// TODO: make this a transaction
	if err := s.userStore.Insert(ctx, user); err != nil {
		return auth.User{}, err
	}
	if err := s.passwordStore.Insert(ctx, password); err != nil {
		return auth.User{}, err
	}

	// TODO: send user verification

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, creds auth.Credentials) (auth.User, auth.Tokens, error) {
	now := time.Now()

	user, err := s.userStore.FindByName(ctx, creds.Username)
	if err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	password, err := s.passwordStore.FindByUsername(ctx, creds.Username)
	if err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	passwordAttempt, err := s.makePassword(password.Salt.Data, creds, passwordOptions{
		digestType:    password.DigestType,
		hashKeyLength: password.KeyLength,
		hashRounds:    password.Iterations,
	})
	if err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	if !bytes.Equal(password.HashedPassword.Data, passwordAttempt.HashedPassword.Data) {
		return auth.User{}, auth.Tokens{}, common.NewErr("invalid password", common.ErrCodeBadRequest)
	}

	user, tokens, err := s.makeSession(ctx, user.ID, primitive.NilObjectID, now)
	if err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	return user, tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, userID primitive.ObjectID) error {
	if err := s.userStore.ClearSessions(ctx, userID); err != nil {
		return err
	}

	if err := s.refreshTokenStore.DeleteByUserID(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) RefreshAccess(ctx context.Context, refreshToken auth.RefreshToken) (auth.User, auth.Tokens, error) {
	now := time.Now()

	if err := s.refreshTokenStore.Consume(ctx, refreshToken.SessionID); err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	user, tokens, err := s.makeSession(ctx, refreshToken.UserID, refreshToken.SessionID, now)
	if err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	return user, tokens, nil
}

func (s *AuthService) ParseToken(payload string, claims jwt.Claims) error {
	_, err := jwt.ParseWithClaims(payload, claims, s.tokenKeyFunc)
	if err != nil {
		if err.Error() == "signature is invalid" {
			return auth.ErrInvalidSignature
		}
		return auth.ErrMalformedToken
	}
	if claims, ok := claims.(common.Validator); ok {
		if err := claims.Validate(); err != nil {
			return auth.ErrInvalidToken(err)
		}
	}
	return nil
}

func (s *AuthService) SignToken(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func (s *AuthService) tokenKeyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("%v is an unsupported signing method", token.Header["alg"])
	}
	return []byte(s.jwtSecret), nil
}

func (s *AuthService) makeSession(ctx context.Context, userID primitive.ObjectID, prevSessionID primitive.ObjectID, now time.Time) (auth.User, auth.Tokens, error) {
	sessionID := primitive.NewObjectID()

	accessToken := s.makeAccessToken(sessionID, userID, now)
	refreshToken := s.makeRefreshToken(accessToken)

	// TODO: make this a transaction
	if err := s.refreshTokenStore.Insert(ctx, refreshToken); err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	user, err := s.userStore.AddSession(ctx, userID, sessionID)
	if err != nil {
		return auth.User{}, auth.Tokens{}, err
	}

	if !prevSessionID.IsZero() {
		user, err = s.userStore.RemoveSession(ctx, userID, prevSessionID)
		if err != nil {
			return auth.User{}, auth.Tokens{}, err
		}
	}

	return user, auth.Tokens{accessToken, refreshToken}, nil
}

func (s *AuthService) makeAccessToken(sessionID, userID primitive.ObjectID, issuedAt time.Time) auth.AccessToken {
	return auth.AccessToken{
		SessionID: sessionID,
		UserID:    userID,
		Issuer:    s.jwtIssuer,
		Audience:  []string{audAPIAdminV1},
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt.Add(s.jwtDurationAccess),
	}
}

func (s *AuthService) makeRefreshToken(accessToken auth.AccessToken) auth.RefreshToken {
	return auth.RefreshToken{
		AccessToken: auth.AccessToken{
			SessionID: accessToken.SessionID,
			UserID:    accessToken.UserID,
			Issuer:    accessToken.Issuer,
			Audience:  accessToken.Audience,
			IssuedAt:  accessToken.IssuedAt,
			ExpiresAt: accessToken.IssuedAt.Add(s.jwtDurationRefresh),
		},
	}
}

type passwordOptions struct {
	digestType    string
	hashKeyLength int
	hashRounds    int
}

func (s *AuthService) makePassword(salt []byte, credentials auth.Credentials, opts passwordOptions) (auth.Password, error) {
	digestType := opts.digestType
	if digestType == auth.DigestTypeEmpty {
		digestType = auth.DigestTypeSHA256
	}

	hashKeyLength := opts.hashKeyLength
	if hashKeyLength == 0 {
		hashKeyLength = defaultHashKeyLength
	}

	hashRounds := opts.hashRounds
	if hashRounds == 0 {
		hashRounds = defaultHashRounds
	}

	password := auth.Password{
		Username: credentials.Username,
		Salt:     primitive.Binary{Data: salt},
		HashedPassword: primitive.Binary{Data: []byte(hex.EncodeToString(pbkdf2.Key(
			[]byte(credentials.Password),
			append(s.passwordSalt, salt...),
			hashRounds,
			hashKeyLength,
			auth.DigestHash(digestType),
		)))},
		Iterations: hashRounds,
		KeyLength:  hashKeyLength,
		DigestType: digestType,
	}

	if err := password.Validate(); err != nil {
		return auth.Password{}, common.WrapErr(fmt.Errorf("failed to make password: %s", err), common.ErrCodeServer)
	}
	return password, nil
}
