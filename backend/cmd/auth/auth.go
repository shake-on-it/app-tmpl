package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/core"
	"github.com/shake-on-it/app-tmpl/backend/core/mongodb"

	"github.com/joho/godotenv"
	"github.com/kr/pretty"
	cli "github.com/urfave/cli/v2"
)

func main() {
	cmd := &cli.App{
		Name:  "auth",
		Usage: "manage application auth",
		Commands: []*cli.Command{
			{
				Name:    "users",
				Aliases: []string{"user"},
				Usage:   "manage application users",
				Subcommands: []*cli.Command{
					{
						Name:   "add",
						Usage:  "add a new application user",
						Flags:  addUserFlags,
						Action: addUser,
					},
				},
			},
		},
	}

	if err := cmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var (
	addUserFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     "username",
			Usage:    "the new user's username",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password",
			Usage:    "the new user's password",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "email",
			Usage:    "the new user's email",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "the new user's type",
		},
		&cli.StringFlag{
			Name:  "status",
			Usage: "the new user's status",
		},
		&cli.StringFlag{
			Name:  "mongo_uri",
			Usage: "the mongodb uri to connect to",
		},
		&cli.StringFlag{
			Name:  "salt",
			Usage: "the salt to use with new password",
		},
	}
)

func addUser(cliCtx *cli.Context) error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	mongoURI := cliCtx.String("mongo_uri")
	if mongoURI == "" {
		mongoURI = os.Getenv("app_mongodb_url")
	}
	if mongoURI == "" {
		return errors.New("must specify mongo uri")
	}

	salt := cliCtx.String("salt")
	if salt == "" {
		salt = os.Getenv("auth_password_salt")
	}
	if salt == "" {
		return errors.New("must specify salt")
	}

	logger, err := common.NewLogger("auth", common.LoggerOptionsDev)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.TimeoutServerOp)
	defer cancel()

	mongoProvider := mongodb.NewProvider(mongoURI, logger)
	if err := mongoProvider.Setup(ctx); err != nil {
		return err
	}

	userStore, err := core.NewUserStore(mongoProvider.Client())
	if err != nil {
		return err
	}

	passwordStore, err := core.NewPasswordStore(mongoProvider.Client())
	if err != nil {
		return err
	}

	refreshTokenStore, err := core.NewRefreshTokenStore(mongoProvider.Client())
	if err != nil {
		return err
	}

	authService := core.NewAuthService(
		common.Config{Auth: common.AuthConfig{PasswordSalt: salt}},
		userStore,
		passwordStore,
		refreshTokenStore,
	)

	user, err := authService.CreateUser(ctx, auth.Registration{
		auth.Credentials{
			Username: cliCtx.String("username"),
			Password: cliCtx.String("password"),
		},
		cliCtx.String("email"),
	})
	if err != nil {
		return err
	}

	logger.Infof("successfully created user:\n%# v\n", pretty.Formatter(user))

	return nil
}
