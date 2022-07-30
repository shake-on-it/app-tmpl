package common

import (
	"time"
)

const (
	TimeoutServerOp = time.Minute
)

var (
	ServerEnv       string
	ServerGitHash   string
	ServerBuildTime string
)

func ServerVersion() (string, string, string) {
	env, gitHash, buildTime := ServerEnv, ServerGitHash, ServerBuildTime

	if env == "" {
		env = "n/a"
	}
	if gitHash == "" {
		gitHash = "n/a"
	}
	if buildTime == "" {
		buildTime = "n/a"
	}
	return env, gitHash, buildTime
}
