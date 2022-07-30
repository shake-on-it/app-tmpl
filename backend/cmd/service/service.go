package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/api/server"
	"github.com/shake-on-it/app-tmpl/backend/common"

	"github.com/drone/envsubst"
	"go.uber.org/zap"
)

var (
	pathConfig     string
	pathEncryption string
)

func main() {
	flag.StringVar(&pathConfig, "config", "", "config path")
	flag.StringVar(&pathEncryption, "encryption", "", "encryption path")
	flag.Parse()

	config, err := parseConfig()
	if err != nil {
		log.Fatalf("failed to parse server config: %s", err)
	}

	var crypter common.Crypter
	if pathEncryption != "" {
		crypter, err = common.LoadCrypter(pathEncryption)
		if err != nil {
			log.Fatalf("failed to load server crypter: %s", err)
		}
	}

	loggerOpts := common.LoggerOptionsDev
	if config.Env == common.EnvProd {
		loggerOpts = common.LoggerOptionsProd
	}

	logger, err := common.NewLogger("shake_on_it", loggerOpts)
	if err != nil {
		log.Fatalf("failed to make server logger: %s", err)
	}
	logger = logger.With(zap.String("env", config.Env.String()))

	logger.Info("starting up")
	service, err := startService(config, crypter, logger)
	if err != nil {
		logger.Error("failed to start server: %s", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	logger.Info("shutting down")
	if err := service.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown server: %s", err)
		os.Exit(1)
	}
	logger.Info("server shutdown complete")
}

func parseConfig() (common.Config, error) {
	var config common.Config

	dataRaw, err := ioutil.ReadFile(pathConfig)
	if err != nil {
		return common.Config{}, err
	}

	data, err := envsubst.EvalEnv(string(dataRaw))
	if err != nil {
		return common.Config{}, err
	}

	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return common.Config{}, err
	}
	if err := config.Validate(); err != nil {
		return common.Config{}, err
	}
	return config, nil
}

func startService(config common.Config, crypter common.Crypter, logger common.Logger) (*server.Service, error) {
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := server.NewService(config, crypter, logger)

	if err := service.Setup(ctx); err != nil {
		return nil, err
	}

	go service.Start()

	// go health checker

	<-exitCh
	return &service, nil
}
