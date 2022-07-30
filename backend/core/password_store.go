package core

import (
	"context"
	"fmt"

	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/core/mongodb"
	"github.com/shake-on-it/app-tmpl/backend/core/namespaces"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PasswordStore interface {
	FindByUsername(ctx context.Context, username string) (auth.Password, error)

	Insert(ctx context.Context, password auth.Password) error

	UpdatePassword(ctx context.Context, username string, salt, hashedPassword primitive.Binary) error
}

func NewPasswordStore(client *mongo.Client) (PasswordStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.TimeoutServerOp)
	defer cancel()

	coll, err := mongodb.NewColl(ctx, client, namespaces.DBAuth, namespaces.CollPasswords, mongodb.Index{
		Unique: true,
		Key: mongodb.NewIndexKey(
			mongodb.IndexField{namespaces.FieldUsername, 1}),
	})
	if err != nil {
		return nil, err
	}

	return &passwordStore{coll}, nil
}

type passwordStore struct {
	coll *mongo.Collection
}

func (s *passwordStore) FindByUsername(ctx context.Context, username string) (auth.Password, error) {
	var password auth.Password
	if err := s.coll.FindOne(ctx, bson.D{{namespaces.FieldUsername, username}}).Decode(&password); err != nil {
		if err == mongo.ErrNoDocuments {
			return auth.Password{}, common.NewErr("must register first", common.ErrCodeNotFound)
		}
		return auth.Password{}, common.WrapErr(fmt.Errorf("failed to find password: %s", err), common.ErrCodeServer)
	}
	return password, nil
}

func (s *passwordStore) Insert(ctx context.Context, password auth.Password) error {
	if _, err := s.coll.InsertOne(ctx, password); err != nil {
		return common.WrapErr(fmt.Errorf("failed to create password: %s", err), common.ErrCodeServer)
	}
	return nil
}

func (s *passwordStore) UpdatePassword(ctx context.Context, username string, salt, hashedPassword primitive.Binary) error {
	if _, err := s.coll.UpdateOne(
		ctx,
		bson.D{{namespaces.FieldUsername, username}},
		bson.D{
			{namespaces.FieldSalt, salt},
			{namespaces.FieldHashedPassword, hashedPassword},
		},
	); err != nil {
		return common.WrapErr(fmt.Errorf("failed to update password: %s", err), common.ErrCodeServer)
	}
	return nil
}
