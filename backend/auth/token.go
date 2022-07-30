package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/common"

	jwt "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CookieAccessToken  = "access-token"
	CookieRefreshToken = "refresh-token"
	CookieUserToken    = "user-token"
)

var (
	ErrInvalidSession   = common.NewErr("invalid session", common.ErrCodeInvalidAuth)
	ErrInvalidSignature = common.NewErr("invalid signature", common.ErrCodeInvalidAuth)
	ErrMalformedCookie  = common.NewErr("cookie is malformed", common.ErrCodeBadRequest)
	ErrMalformedToken   = common.NewErr("token is malformed", common.ErrCodeInvalidAuth)
	ErrMustAuthenticate = common.NewErr("must authenticate", common.ErrCodeInvalidAuth)
)

func ErrInvalidToken(err error) error {
	return common.WrapErr(fmt.Errorf("invalid token: %s", err), common.ErrCodeInvalidAuth)
}

type Tokens struct {
	AccessToken  AccessToken
	RefreshToken RefreshToken
}

type AccessToken struct {
	SessionID primitive.ObjectID `bson:"_id"`
	UserID    primitive.ObjectID `bson:"sub"`
	Issuer    string             `bson:"iss"`
	Audience  []string           `bson:"aud"`
	IssuedAt  time.Time          `bson:"iat"`
	ExpiresAt time.Time          `bson:"exp"`
}

type RefreshToken struct {
	AccessToken `bson:",inline"`
	Consumed    bool `bson:"consumed"`
}

func (t *AccessToken) Valid() error {
	return nil
}

func (t *AccessToken) Validate() error {
	if t.SessionID.IsZero() {
		return errors.New("token needs session")
	}
	if t.UserID.IsZero() {
		return errors.New("token needs user")
	}
	if time.Now().After(t.ExpiresAt) {
		return errors.New("token is expired")
	}
	return nil
}

func (t AccessToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(jwt.RegisteredClaims{
		ID:        t.SessionID.Hex(),
		Subject:   t.UserID.Hex(),
		Issuer:    t.Issuer,
		Audience:  t.Audience,
		IssuedAt:  jwt.NewNumericDate(t.IssuedAt),
		ExpiresAt: jwt.NewNumericDate(t.ExpiresAt),
	})
}

func (t *AccessToken) UnmarshalJSON(data []byte) error {
	var claims jwt.RegisteredClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return err
	}

	sessionID, err := primitive.ObjectIDFromHex(claims.ID)
	if err != nil {
		return err
	}

	userID, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		return err
	}

	t.SessionID = sessionID
	t.UserID = userID
	t.Issuer = claims.Issuer
	t.Audience = claims.Audience
	t.IssuedAt = claims.IssuedAt.Time
	t.ExpiresAt = claims.ExpiresAt.Time

	return nil
}
