package common

import (
	"fmt"
	"strings"
	"time"
)

type Env string

const (
	EnvLocal = ""
	EnvDev   = "dev"
	EnvProd  = "prod"
	EnvTest  = "test"
)

func (e Env) String() string {
	if e == EnvLocal {
		return "local"
	}
	return string(e)
}

type Config struct {
	Env    Env          `json:"env"`
	API    APIConfig    `json:"api"`
	Auth   AuthConfig   `json:"auth"`
	DB     DBConfig     `json:"db"`
	Server ServerConfig `json:"server"`
}

func (c *Config) Validate() error {
	if err := c.API.validate(); err != nil {
		return err
	}
	if err := c.Auth.validate(); err != nil {
		return err
	}
	if err := c.Server.validate(); err != nil {
		return err
	}
	return nil
}

const (
	defaultAPIRequestLimit = 60_000
)

type APIConfig struct {
	CORSOrigins []string `json:"cors_origins"`

	RequestLimit       int `json:"request_limit"`
	RequestTimeoutSecs int `json:"request_timeout_secs"`
}

func (c APIConfig) RequestTimeout() time.Duration {
	return time.Duration(c.RequestTimeoutSecs) * time.Second
}

func (c *APIConfig) validate() error {
	if c.RequestLimit == 0 {
		c.RequestLimit = defaultAPIRequestLimit
	}
	return nil
}

const (
	defaultAccessTokenExpirySecs  = 5 * 60
	defaultRefreshTokenExpiryDays = 30
)

type AuthConfig struct {
	JWTSecret              string `json:"jwt_secret"`
	PasswordSalt           string `json:"password_salt"`
	AccessTokenExpirySecs  int    `json:"access_token_expiry_secs"`
	RefreshTokenExpiryDays int    `json:"refresh_token_expiry_days"`
}

func (c *AuthConfig) validate() error {
	if c.AccessTokenExpirySecs == 0 {
		c.AccessTokenExpirySecs = defaultAccessTokenExpirySecs
	}
	if c.RefreshTokenExpiryDays == 0 {
		c.RefreshTokenExpiryDays = defaultRefreshTokenExpiryDays
	}
	return nil
}

func (c AuthConfig) AccessTokenExpiry() time.Duration {
	return time.Duration(c.AccessTokenExpirySecs) * time.Second
}

func (c AuthConfig) RefreshTokenExpiry() time.Duration {
	return time.Duration(c.RefreshTokenExpiryDays) * 24 * time.Hour
}

type DBConfig struct {
	URI    string `json:"uri"`
}

type ServerConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	SSLEnabled bool   `json:"ssl_enabled"`
	BaseURL    string `json:"-"`
}

func (c *ServerConfig) validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("server port is out of range range: %d", c.Port)
	}
	c.BaseURL = c.baseURL()
	return nil
}

func (c ServerConfig) baseURL() string {
	var sb strings.Builder

	sb.WriteString("http")
	if c.SSLEnabled {
		sb.WriteString("s")
	}
	sb.WriteString("://")

	host := c.Host
	if host == "" {
		host = "localhost"
	}
	sb.WriteString(host)

	if c.Port != 0 {
		sb.WriteString(fmt.Sprintf(":%d", c.Port))
	}

	return sb.String()
}
