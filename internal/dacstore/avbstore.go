package dacstore

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AvailabilityStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (a AvailabilityStore) Client() *mongo.Client         { return a.client }
func (a AvailabilityStore) Collection() *mongo.Collection { return a.collection }

func newAvailabilityStore(client *mongo.Client, dbname, collectionName string) *AvailabilityStore {
	collection := client.Database(dbname).Collection(collectionName)
	return &AvailabilityStore{client: client, collection: collection}
}

func (a AvailabilityStore) FetchByNewest(ctx context.Context) (models.Modler, error) {
	var avb models.Availability

	filter := bson.M{"$and": bson.A{
		bson.M{"newest": true},
		bson.M{"day": bson.M{"$in": bson.A{0, 1, 2, 3, 4, 5, 6}}},
	}}

	err := a.collection.FindOne(ctx, filter).Decode(&avb)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NewNoResultErr("newest", "Availability", err)
		}
		logger.MustDebug(fmt.Sprintf("error occurred whild fetching newest availability, %v", err))
		return nil, err
	}

	return &avb, nil
}
