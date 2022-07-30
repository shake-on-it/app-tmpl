package v1_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/common/test"
	"github.com/shake-on-it/app-tmpl/backend/common/test/assert"
)

func TestAuth(t *testing.T) {
	t.Run("should fail to log in when no access token is present", func(t *testing.T) {
		th := test.NewHarness(t)
		defer th.Close()

		res, err := th.Do(test.Request{
			Path: "/api/admin/v1/user",
			Auth: true,
		})
		assert.Nil(t, err)
		assert.Nil(t, res.Is(http.StatusUnauthorized))

		assert.Equal(t, res.Err(), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "must authenticate",
		})
	})

	t.Run("should fail to log in if access token is expired", func(t *testing.T) {
		th := test.NewHarnessWithOptions(t, test.HarnessOptions{
			Config: common.Config{
				Auth: common.AuthConfig{
					AccessTokenExpirySecs: 2,
				},
			},
		})
		defer th.Close()

		assert.Nil(t, th.Login())

		time.Sleep(3 * time.Second)

		res, err := th.Do(test.Request{
			Path: "/api/admin/v1/user",
			Auth: true,
		})
		assert.Nil(t, err)
		assert.Nil(t, res.Is(http.StatusUnauthorized))

		assert.Equal(t, res.Err(), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "invalid token: token is expired",
		})
	})

	t.Run("should be able to log in and get system status", func(t *testing.T) {
		th := test.NewHarness(t)
		defer th.Close()

		assert.Nil(t, th.Login())

		res, err := th.Do(test.Request{
			Path: "/api/admin/v1/system/status",
			Auth: true,
		})
		assert.Nil(t, err)
		assert.Nil(t, res.Is(http.StatusOK))

		var out string
		assert.Nil(t, res.Decode(&out))
		assert.Equal(t, out, "you are authenticated")
	})
}
