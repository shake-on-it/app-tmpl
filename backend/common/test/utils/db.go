package utils

import (
	"context"
	"os"
	"testing"

	"github.com/shake-on-it/app-tmpl/backend/core/mongodb"
	"github.com/shake-on-it/app-tmpl/backend/core/namespaces"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	envMongoDBURI = "MONGODB_URI"
	envNoSkip     = "NO_SKIP"

	defaultMongoDBURI = "mongodb://localhost:27017"
)

var (
	mongodbChecked   = false
	mongodbConnected = false
	mongodbProvider  mongodb.Provider
)

func MongoProvider() mongodb.Provider {
	if mongodbProvider == nil {
		panic("must call u.SkipUnlessMongoRunning(t) in test function")
	}
	return mongodbProvider
}

func MongoURI() string {
	if !mongodbChecked {
		panic("must call u.SkipUnlessMongoRunning(t) in test function")
	}
	if uri := os.Getenv(envMongoDBURI); uri != "" {
		return uri
	}
	return defaultMongoDBURI
}

var SkipUnlessMongoRunning = func() func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		setupMongo()
		t.Cleanup(teardownMongo)

		mongodbChecked = true
		uri := MongoURI()

		mongodbProvider = mongodb.NewProvider(MongoURI(), NewLogger(t))

		if err := mongodbProvider.Setup(context.Background()); err != nil {
			MustSkip(t, "failed to connect to mongodb at "+uri)
			return
		}
		mongodbConnected = true
	}
}()

var (
	originalDatabases   = map[string]string{}
	originalCollections = map[string]string{}

	randomizedDatabases   = map[string]string{}
	randomizedCollections = map[string]string{}

	testNamespaces = map[*string][]*string{}
)

func setupMongo() {
	for _, ns := range namespaces.Registry {
		origDB, origColl := ns.Database, ns.Collection
		testNamespaces[origDB] = append(testNamespaces[origDB], origColl)
	}

	for db, colls := range testNamespaces {
		newDB := primitive.NewObjectID().Hex()
		randomizedDatabases[*db] = newDB
		originalDatabases[newDB] = *db
		*db = newDB

		for _, coll := range colls {
			newColl := primitive.NewObjectID().Hex()
			randomizedCollections[*coll] = newColl
			originalCollections[newColl] = *coll
			*coll = newColl
		}
	}
}

func teardownMongo() {
	if !mongodbConnected {
		return
	}

	defer mongodbProvider.Close(context.Background())

	for _, ns := range namespaces.Registry {

		mongodbProvider.Client().
			Database(*ns.Database).
			Collection(*ns.Collection).
			Drop(context.Background())
	}

	for _, ns := range namespaces.Registry {
		if db, ok := originalDatabases[*ns.Database]; ok {
			*ns.Database = db
		}
		if coll, ok := originalCollections[*ns.Collection]; ok {
			*ns.Collection = coll
		}
	}
}
