package dacstore

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VisitorStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (v VisitorStore) Client() *mongo.Client         { return v.client }
func (v VisitorStore) Collection() *mongo.Collection { return v.collection }

func newVisitorStore(client *mongo.Client, dbname, collName string) *VisitorStore {
	collection := client.Database(dbname).Collection(collName)
	return &VisitorStore{client: client, collection: collection}
}

func (v VisitorStore) Insert(ctx context.Context, vis models.Visitor) (primitive.ObjectID, error) {
	result, err := v.collection.InsertOne(ctx, vis)
	if err != nil {
		insertErr := fmt.Errorf("error inserting Visitor record:: %w", err)
		logger.MustDebug(fmt.Sprintf("error inserting visitor record:: %v", err))
		return primitive.NilObjectID, insertErr
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (v VisitorStore) FetchVisitor(ctx context.Context, ip string) (models.Modler, error) {
	var vis models.Visitor
	filter := bson.M{"address": ip}
	err := v.collection.FindOne(ctx, filter).Decode(&vis)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NewNoResultErr(ip, "visitor", err)
		}
		return nil, err
	}
	return &vis, nil
}

func (v VisitorStore) UpdateVisitorCount(ctx context.Context, ip string) (int64, int64, error) {
	filter := bson.M{"address": ip}
	update := bson.M{"$inc": bson.M{"visit_count": 1}}
	result, err := v.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return 0, 0, err
	}

	return result.MatchedCount, result.ModifiedCount, nil
}
