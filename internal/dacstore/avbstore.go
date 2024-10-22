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

func (a AvailabilityStore) FetchByNewest(ctx context.Context) ([]models.Availability, error) {
	var avbs []models.Availability

	filter := bson.M{"$and": bson.A{
		bson.M{"newest": true},
		bson.M{"day": bson.M{"$in": bson.A{0, 1, 2, 3, 4, 5, 6}}},
	}}

	cur, err := a.collection.Find(ctx, filter)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error occurrd while fetching availability:: %v", err))
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var avb models.Availability
		if err := cur.Decode(&avb); err != nil {
			logger.MustDebug(fmt.Sprintf("error decoding availability to json:: %v", err))
			continue
		}
		avbs = append(avbs, avb)
	}

	if err := cur.Err(); err != nil {
		logger.MustDebug(fmt.Sprintf("error occurred with availability cursor:: %v", err))
		return nil, err
	}

	return avbs, nil
}
