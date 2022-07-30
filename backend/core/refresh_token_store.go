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

type RefreshTokenStore interface {
	Check(ctx context.Context, id primitive.ObjectID) (bool, error)

	Insert(ctx context.Context, refreshToken auth.RefreshToken) error

	Consume(ctx context.Context, id primitive.ObjectID) error

	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

func NewRefreshTokenStore(client *mongo.Client) (RefreshTokenStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.TimeoutServerOp)
	defer cancel()

	coll, err := mongodb.NewColl(ctx, client, namespaces.DBAuth, namespaces.CollRefreshTokens, mongodb.Index{
		Key: mongodb.NewIndexKey(
			mongodb.IndexField{namespaces.FieldSub, 1}),
	})
	if err != nil {
		return nil, err
	}

	return &refreshTokenStore{coll}, nil
}

type refreshTokenStore struct {
	coll *mongo.Collection
}

func (s *refreshTokenStore) Check(ctx context.Context, id primitive.ObjectID) (bool, error) {
	res := s.coll.FindOne(ctx, bson.D{{namespaces.FieldID, id}})
	if err := res.Err(); err != nil {
		return false, common.WrapErr(fmt.Errorf("failed to find refresh token: %s", err), common.ErrCodeServer)
	}

	var refreshToken auth.RefreshToken
	res.Decode(&refreshToken)
	return !refreshToken.Consumed, nil
}

func (s *refreshTokenStore) Insert(ctx context.Context, refreshToken auth.RefreshToken) error {
	if _, err := s.coll.InsertOne(ctx, refreshToken); err != nil {
		return common.WrapErr(fmt.Errorf("failed to create session: %s", err), common.ErrCodeServer)
	}
	return nil
}

func (s *refreshTokenStore) Consume(ctx context.Context, id primitive.ObjectID) error {
	res, err := s.coll.UpdateOne(
		ctx,
		bson.D{
			{namespaces.FieldID, id},
			{namespaces.FieldConsumed, false},
		},
		bson.D{{"$set", bson.D{
			{namespaces.FieldConsumed, true},
		}}},
	)
	if err != nil {
		return common.WrapErr(fmt.Errorf("failed to refresh session: %s", err), common.ErrCodeServer)
	}
	if res.MatchedCount == 0 {
		return common.NewErr("session has expired", common.ErrCodeInvalidAuth)
	}
	return nil
}

func (s *refreshTokenStore) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	if _, err := s.coll.DeleteMany(ctx, bson.D{{namespaces.FieldSub, userID}}); err != nil {
		return common.WrapErr(fmt.Errorf("failed to delete sessions: %s", err), common.ErrCodeServer)
	}
	return nil
}
