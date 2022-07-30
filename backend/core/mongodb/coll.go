package mongodb

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewColl(ctx context.Context, client *mongo.Client, db, coll string, indexes ...Index) (*mongo.Collection, error) {
	for i := range indexes {
		if indexes[i].Name != "" {
			continue
		}
		var sb strings.Builder
		for j, field := range indexes[i].Key.Fields {
			if j > 0 {
				sb.WriteString("_")
			}
			sb.WriteString(fmt.Sprintf("%s_%v", field.Name, field.Value))
		}
		indexes[i].Name = sb.String()
	}

	res := client.
		Database(db).
		RunCommand(ctx, bson.D{
			{"createIndexes", coll},
			{"indexes", indexes},
		})
	if err := res.Err(); err != nil {
		return nil, err
	}

	return client.Database(db).Collection(coll), nil
}

type Index struct {
	Name                    string   `bson:"name"`
	Key                     IndexKey `bson:"key"`
	Unique                  bool     `bson:"unique"`
	ExpireAfterSeconds      int      `bson:"expire_after_seconds,omitempty"`
	PartialFilterExpression bson.D   `bson:"partial_filter_expression,omitempty"`
}

type IndexField struct {
	Name  string
	Value interface{}
}

func NewIndexKey(fields ...IndexField) IndexKey {
	return IndexKey{fields}
}

type IndexKey struct {
	Fields []IndexField
}

func (k IndexKey) MarshalBSON() ([]byte, error) {
	d := make(bson.D, 0, len(k.Fields))
	for _, f := range k.Fields {
		d = append(d, bson.E{f.Name, f.Value})
	}
	return bson.Marshal(d)
}

func CreateIndexes(ctx context.Context, client *mongo.Client, db, coll string, indexes []Index) error {
	for i := range indexes {
		if indexes[i].Name != "" {
			continue
		}
		var sb strings.Builder
		for j, field := range indexes[i].Key.Fields {
			if j > 0 {
				sb.WriteString("_")
			}
			sb.WriteString(fmt.Sprintf("%s_%v", field.Name, field.Value))
		}
		indexes[i].Name = sb.String()
	}

	return client.
		Database(db).
		RunCommand(ctx, bson.D{
			{"createIndexes", coll},
			{"indexes", indexes},
		}).
		Err()
}
