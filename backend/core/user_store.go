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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserStore interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (auth.User, error)
	FindByName(ctx context.Context, name string) (auth.User, error)

	Insert(ctx context.Context, user auth.User) error

	AddSession(ctx context.Context, id, sessionID primitive.ObjectID) (auth.User, error)
	RemoveSession(ctx context.Context, id, sessionID primitive.ObjectID) (auth.User, error)
	ClearSessions(ctx context.Context, id primitive.ObjectID) error
}

func NewUserStore(client *mongo.Client) (UserStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.TimeoutServerOp)
	defer cancel()

	coll, err := mongodb.NewColl(ctx, client, namespaces.DBAuth, namespaces.CollUsers, mongodb.Index{
		Unique: true,
		Key: mongodb.NewIndexKey(
			mongodb.IndexField{namespaces.FieldName, 1}),
	})
	if err != nil {
		return nil, err
	}

	return &userStore{coll}, nil
}

type userStore struct {
	coll *mongo.Collection
}

func (s *userStore) FindByID(ctx context.Context, id primitive.ObjectID) (auth.User, error) {
	var user auth.User
	if err := s.coll.FindOne(ctx, bson.D{{namespaces.FieldID, id}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return auth.User{}, common.NewErr("cannot find user", common.ErrCodeNotFound)
		}
		return auth.User{}, common.WrapErr(fmt.Errorf("failed to find user: %s", err), common.ErrCodeServer)
	}
	return user, nil
}

func (s *userStore) FindByName(ctx context.Context, name string) (auth.User, error) {
	var user auth.User
	if err := s.coll.FindOne(ctx, bson.D{{namespaces.FieldName, name}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return auth.User{}, common.NewErr("cannot find user", common.ErrCodeNotFound)
		}
		return auth.User{}, common.WrapErr(fmt.Errorf("failed to find user: %s", err), common.ErrCodeServer)
	}
	return user, nil
}

func (s *userStore) Insert(ctx context.Context, user auth.User) error {
	if _, err := s.coll.InsertOne(ctx, user); err != nil {
		return common.WrapErr(fmt.Errorf("failed to create user: %s", err), common.ErrCodeServer)
	}
	return nil
}

func (s *userStore) AddSession(ctx context.Context, id, sessionID primitive.ObjectID) (auth.User, error) {
	res := s.coll.FindOneAndUpdate(
		ctx,
		bson.D{{namespaces.FieldID, id}},
		bson.D{{"$push", bson.D{
			{namespaces.FieldSessions, sessionID},
		}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if err := res.Err(); err != nil {
		return auth.User{}, common.WrapErr(fmt.Errorf("failed to add user session: %s", err), common.ErrCodeServer)
	}

	var user auth.User
	res.Decode(&user)
	return user, nil
}

func (s *userStore) RemoveSession(ctx context.Context, id, sessionID primitive.ObjectID) (auth.User, error) {
	res := s.coll.FindOneAndUpdate(
		ctx,
		bson.D{{namespaces.FieldID, id}},
		bson.D{{"$pull", bson.D{
			{namespaces.FieldSessions, sessionID},
		}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if err := res.Err(); err != nil {
		return auth.User{}, common.WrapErr(fmt.Errorf("failed to remove user session: %s", err), common.ErrCodeServer)
	}

	var user auth.User
	res.Decode(&user)
	return user, nil
}

func (s *userStore) ClearSessions(ctx context.Context, id primitive.ObjectID) error {
	if _, err := s.coll.UpdateOne(
		ctx,
		bson.D{{namespaces.FieldID, id}},
		bson.D{{"$unset", bson.D{{namespaces.FieldSessions, 1}}}},
	); err != nil {
		return common.WrapErr(fmt.Errorf("failed to clear user sessions: %s", err), common.ErrCodeServer)
	}
	return nil
}
