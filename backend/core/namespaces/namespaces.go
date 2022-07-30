package namespaces

var (
	DBApp    = "tmpl_app"
	CollBets = "bets"

	DBAuth            = "tmpl_auth"
	CollRefreshTokens = "refresh_tokens"
	CollPasswords     = "passwords"
	CollUsers         = "users"
)

type Namespace struct {
	Database   *string
	Collection *string
}

var (
	Registry = []Namespace{
		{&DBApp, &CollBets},
		{&DBAuth, &CollRefreshTokens},
		{&DBAuth, &CollPasswords},
		{&DBAuth, &CollUsers},
	}
)

const (
	FieldID       = "_id"
	FieldName     = "name"
	FieldSessions = "sessions"

	FieldConsumed = "consumed"
	FieldSub      = "sub"

	FieldUsername       = "username"
	FieldSalt           = "salt"
	FieldHashedPassword = "hashed_password"
)
