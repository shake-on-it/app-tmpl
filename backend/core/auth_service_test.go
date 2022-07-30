package core

import (
	"context"
	"testing"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/common/test/assert"
	u "github.com/shake-on-it/app-tmpl/backend/common/test/utils"
	"github.com/shake-on-it/app-tmpl/backend/core/namespaces"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuthService(t *testing.T) {
	u.SkipUnlessMongoRunning(t)

	client := u.MongoProvider().Client()

	userStore, err := NewUserStore(client)
	assert.Nil(t, err)

	passwordStore, err := NewPasswordStore(client)
	assert.Nil(t, err)

	refreshTokenStore, err := NewRefreshTokenStore(client)
	assert.Nil(t, err)

	s := NewAuthService(
		common.Config{
			Auth: common.AuthConfig{
				AccessTokenExpirySecs:  3600,
				RefreshTokenExpiryDays: 1,
				PasswordSalt:           "abcdefghijkl",
			},
			Server: common.ServerConfig{
				BaseURL: "http://localhost",
			},
		},
		userStore,
		passwordStore,
		refreshTokenStore,
	)

	creds := auth.Credentials{
		Username: "username",
		Password: "password",
	}

	t.Run("should successfully create user and password", func(t *testing.T) {
		user, err := s.CreateUser(context.Background(), auth.Registration{
			Credentials: creds,
			Email:       "email@domain.com",
		})
		assert.Nil(t, err)

		assert.Equal(t, user.Name, creds.Username)
		assert.Equal(t, user.Email, "email@domain.com")
		assert.Equal(t, user.Type, auth.UserTypeGuest)
		assert.Equal(t, user.Status, auth.UserStatusUnverified)
		assert.Equal(t, len(user.Sessions), 0)

		password, err := passwordStore.FindByUsername(context.Background(), creds.Username)
		assert.Nil(t, err)

		assert.Equal(t, password.Username, creds.Username)
		assert.Equal(t, password.DigestType, auth.DigestTypeSHA256)
		assert.Equal(t, password.Iterations, defaultHashRounds)
		assert.Equal(t, password.KeyLength, defaultHashKeyLength)

		t.Run("and login with those credentials", func(t *testing.T) {
			now := time.Now()

			user, tokens, err := s.Login(context.Background(), creds)
			assert.Nil(t, err)

			assert.Equal(t, user.Name, creds.Username)

			assert.Equal(t, tokens.AccessToken.SessionID, user.Sessions[0])
			assert.Equal(t, tokens.AccessToken.UserID, user.ID)
			assert.Equal(t, tokens.AccessToken.Issuer, "http://localhost")
			assert.Equal(t, tokens.AccessToken.Audience, []string{"api/admin/v1"})
			assert.True(t, tokens.AccessToken.IssuedAt.After(now))

			assert.Equal(t, tokens.AccessToken.SessionID, tokens.RefreshToken.SessionID)
			assert.Equal(t, tokens.AccessToken.UserID, tokens.RefreshToken.UserID)
			assert.Equal(t, tokens.AccessToken.Issuer, tokens.RefreshToken.Issuer)
			assert.Equal(t, tokens.AccessToken.Audience, tokens.RefreshToken.Audience)
			assert.Equal(t, tokens.AccessToken.IssuedAt, tokens.RefreshToken.IssuedAt)

			assert.Equal(t, tokens.AccessToken.ExpiresAt, tokens.AccessToken.IssuedAt.Add(time.Hour))
			assert.Equal(t, tokens.RefreshToken.ExpiresAt, tokens.RefreshToken.IssuedAt.Add(24*time.Hour))

			assert.False(t, tokens.RefreshToken.Consumed)

			t.Run("and refresh the access token", func(t *testing.T) {
				now := time.Now()

				sameUser, newTokens, err := s.RefreshAccess(context.Background(), tokens.RefreshToken)
				assert.Nil(t, err)

				assert.Equal(t, sameUser.ID, user.ID)

				assert.Equal(t, newTokens.AccessToken.SessionID, sameUser.Sessions[0])
				assert.Equal(t, newTokens.AccessToken.UserID, user.ID)
				assert.Equal(t, newTokens.AccessToken.Issuer, "http://localhost")
				assert.Equal(t, newTokens.AccessToken.Audience, []string{"api/admin/v1"})
				assert.True(t, newTokens.AccessToken.IssuedAt.After(now))

				assert.Equal(t, newTokens.AccessToken.SessionID, newTokens.RefreshToken.SessionID)
				assert.Equal(t, newTokens.AccessToken.UserID, newTokens.RefreshToken.UserID)
				assert.Equal(t, newTokens.AccessToken.Issuer, newTokens.RefreshToken.Issuer)
				assert.Equal(t, newTokens.AccessToken.Audience, newTokens.RefreshToken.Audience)
				assert.Equal(t, newTokens.AccessToken.IssuedAt, newTokens.RefreshToken.IssuedAt)

				assert.True(t, newTokens.AccessToken.IssuedAt.After(tokens.AccessToken.IssuedAt))
				assert.True(t, newTokens.RefreshToken.IssuedAt.After(tokens.RefreshToken.IssuedAt))

				assert.Equal(t, newTokens.AccessToken.ExpiresAt, newTokens.AccessToken.IssuedAt.Add(time.Hour))
				assert.Equal(t, newTokens.RefreshToken.ExpiresAt, newTokens.RefreshToken.IssuedAt.Add(24*time.Hour))

				assert.False(t, tokens.RefreshToken.Consumed)

				t.Run("the old refresh token should be consumed", func(t *testing.T) {
					cursor, err := client.
						Database(namespaces.DBAuth).
						Collection(namespaces.CollRefreshTokens).
						Find(
							context.Background(),
							bson.D{{namespaces.FieldSub, user.ID}},
						)
					assert.Nil(t, err)
					defer cursor.Close(context.Background())

					refreshTokensByID := map[primitive.ObjectID]auth.RefreshToken{}
					for cursor.Next(context.Background()) {
						var refreshToken auth.RefreshToken
						assert.Nil(t, cursor.Decode(&refreshToken))
						refreshTokensByID[refreshToken.SessionID] = refreshToken
					}

					assert.True(t, len(refreshTokensByID) == 2)
					assert.True(t, refreshTokensByID[tokens.RefreshToken.SessionID].Consumed)
				})
			})
		})

		t.Run("and logout of those credentials", func(t *testing.T) {
			assert.Nil(t, s.Logout(context.Background(), user.ID))

			user, err := userStore.FindByID(context.Background(), user.ID)
			assert.Nil(t, err)
			assert.True(t, len(user.Sessions) == 0)

			cursor, err := client.
				Database(namespaces.DBAuth).
				Collection(namespaces.CollRefreshTokens).
				Find(
					context.Background(),
					bson.D{{namespaces.FieldSub, user.ID}},
				)
			assert.Nil(t, err)
			defer cursor.Close(context.Background())

			refreshTokensByID := map[primitive.ObjectID]auth.RefreshToken{}
			for cursor.Next(context.Background()) {
				var refreshToken auth.RefreshToken
				assert.Nil(t, cursor.Decode(&refreshToken))
				refreshTokensByID[refreshToken.SessionID] = refreshToken
			}

			assert.True(t, len(refreshTokensByID) == 0)
		})

		t.Run("and login again", func(t *testing.T) {
			now := time.Now()

			user, tokens, err := s.Login(context.Background(), creds)
			assert.Nil(t, err)

			assert.Equal(t, user.Name, creds.Username)

			assert.Equal(t, tokens.AccessToken.SessionID, user.Sessions[0])
			assert.Equal(t, tokens.AccessToken.UserID, user.ID)
			assert.Equal(t, tokens.AccessToken.Issuer, "http://localhost")
			assert.Equal(t, tokens.AccessToken.Audience, []string{"api/admin/v1"})
			assert.True(t, tokens.AccessToken.IssuedAt.After(now))

			assert.Equal(t, tokens.AccessToken.SessionID, tokens.RefreshToken.SessionID)
			assert.Equal(t, tokens.AccessToken.UserID, tokens.RefreshToken.UserID)
			assert.Equal(t, tokens.AccessToken.Issuer, tokens.RefreshToken.Issuer)
			assert.Equal(t, tokens.AccessToken.Audience, tokens.RefreshToken.Audience)
			assert.Equal(t, tokens.AccessToken.IssuedAt, tokens.RefreshToken.IssuedAt)

			assert.Equal(t, tokens.AccessToken.ExpiresAt, tokens.AccessToken.IssuedAt.Add(time.Hour))
			assert.Equal(t, tokens.RefreshToken.ExpiresAt, tokens.RefreshToken.IssuedAt.Add(24*time.Hour))

			assert.False(t, tokens.RefreshToken.Consumed)
		})
	})
}
