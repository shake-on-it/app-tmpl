package mongodb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type Provider interface {
	Client() *mongo.Client

	Setup(ctx context.Context) error
	Close(ctx context.Context)
}

const (
	defaultTimeoutConnect = 30 * time.Second
	defaultTimeoutDial    = 5 * time.Second
	defaultTimeoutSocket  = 30 * time.Second
)

type Settings struct {
	TimeoutConnect time.Duration
	TimeoutDial    time.Duration
	TimeoutSocket  time.Duration
}

func NewProvider(uri string, logger common.Logger) Provider {
	return NewProviderWithSettings(uri, Settings{}, logger)
}

func NewProviderWithSettings(uri string, settings Settings, logger common.Logger) Provider {
	if settings.TimeoutConnect == 0 {
		settings.TimeoutConnect = defaultTimeoutConnect
	}
	if settings.TimeoutDial == 0 {
		settings.TimeoutDial = defaultTimeoutDial
	}
	if settings.TimeoutSocket == 0 {
		settings.TimeoutSocket = defaultTimeoutSocket
	}
	return &provider{uri: uri, settings: settings, logger: logger}
}

type provider struct {
	uri      string
	settings Settings

	client   *mongo.Client
	clientMu sync.Mutex

	logger common.Logger
}

func (p *provider) Setup(ctx context.Context) error {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	if p.client != nil {
		return nil
	}

	if _, err := connstring.Parse(p.uri); err != nil {
		if strings.Contains(err.Error(), "error parsing uri") {
			return errors.New("failed to parse mongodb uri")
		} else if strings.Contains(err.Error(), "_mongodb._tcp") {
			return errors.New("cannot connect to mongodb uri")
		}
	}

	opts := options.Client().
		SetDialer(&net.Dialer{Timeout: p.settings.TimeoutConnect}).
		SetRegistry(bson.DefaultRegistry).
		SetSocketTimeout(p.settings.TimeoutSocket).
		SetConnectTimeout(p.settings.TimeoutConnect).
		SetServerSelectionTimeout(p.settings.TimeoutConnect).
		SetRetryWrites(false)

	ctx, cancel := context.WithTimeout(context.Background(), common.TimeoutServerOp)
	defer cancel()

	client, err := mongo.NewClient(opts.ApplyURI(p.uri))
	if err != nil {
		return fmt.Errorf("failed to build mongodb client: %s", err)
	}

	if client.Connect(context.Background()); err != nil {
		return fmt.Errorf("failed to connect to mongodb: %s", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping mongodb: %s", err)
	}

	p.logger.Info("connected to mongodb")
	p.client = client
	return nil
}

func (p *provider) Client() *mongo.Client {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	if p.client == nil {
		panic("cannot get a mongodb client once the provider is closed")
	}
	return p.client
}

func (p *provider) Close(ctx context.Context) {
	p.clientMu.Lock()
	client := p.client
	p.client = nil
	p.clientMu.Unlock()

	if err := client.Disconnect(ctx); err != nil {
		p.logger.Warnf("failed to disconnect from mongodb client: %s", err)
	}
}
