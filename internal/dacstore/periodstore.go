package dacstore

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PeriodStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (p PeriodStore) Client() *mongo.Client         { return p.client }
func (p PeriodStore) Collection() *mongo.Collection { return p.collection }

func newPeriodStore(client *mongo.Client, dbname, collectionName string) *PeriodStore {
	collection := client.Database(dbname).Collection(collectionName)
	return &PeriodStore{client: client, collection: collection}
}

func (p PeriodStore) Insert(ctx context.Context, prd models.Period) (primitive.ObjectID, error) {
	result, err := p.collection.InsertOne(ctx, prd)
	if err != nil {
		insertErr := fmt.Errorf("error inserting record, %w", err)
		logger.MustDebug("inesrting period failed")
		return primitive.NilObjectID, insertErr
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (p PeriodStore) FetchById(ctx context.Context, id int) (models.Modler, error) {
	var prd models.Period
	filter := bson.M{"_id": id}
	err := p.collection.FindOne(ctx, filter).Decode(&prd)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NewNoResultErr(string(id), "Period", err)
		} else {
			logger.MustDebug(fmt.Sprintf("error occurred while fetching period by id, %s", err))
		}
		return nil, err
	}

	return &prd, nil
}

func (p PeriodStore) FetchActivePeriod(ctx context.Context) (models.Modler, error) {
	var prd models.Period
	filter := bson.M{"IsActive": true}
	err := p.collection.FindOne(ctx, filter).Decode(&prd)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NewNoResultErr("isActive=true", "period", err)
		}
		logger.MustDebug("error occurred while fetching active period")
		return nil, err
	}
	return &prd, nil
}
