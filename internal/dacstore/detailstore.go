package dacstore

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DetailStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (d DetailStore) Client() *mongo.Client         { return d.client }
func (d DetailStore) Collection() *mongo.Collection { return d.collection }

func newDetailStore(ctx context.Context, uri, dbname, collectionName string) (*DetailStore, error) {
	clientOps := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOps)
	if err != nil {
		logger.MustDebug("error connecting to mongo client")
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.MustDebug("could not ping mongo client")
		return nil, err
	}

	collection := client.Database(dbname).Collection(collectionName)
	return &DetailStore{client: client, collection: collection}, nil
}

func (d *DetailStore) Insert(ctx context.Context, dt models.Detail) (primitive.ObjectID, error) {
	result, err := d.collection.InsertOne(ctx, dt)
	if err != nil {
		insertErr := fmt.Errorf("error inserting record %w", err)
		logger.MustDebug("inserting detail failed")
		return primitive.NilObjectID, insertErr
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (d *DetailStore) FetchById(ctx context.Context, id int) (models.Modler, error) {
	var details models.Detail
	filter := bson.M{"_id": id}
	err := d.collection.FindOne(ctx, filter).Decode(&details)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		} else {
			logger.MustDebug("error occurred while fetching detail by id")
		}
		return nil, err
	}

	return &details, nil
}

func (d *DetailStore) FetchByName(ctx context.Context, name string) (models.Modler, error) {
	var details models.Detail
	filter := bson.M{"name": name}
	err := d.collection.FindOne(ctx, filter).Decode(&details)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		} else {
			logger.MustDebug("error occurred while fetching detail by id")
		}
		return nil, err
	}

	return details, nil
}

func (d *DetailStore) Fetch(ctx context.Context) ([]models.Modler, error) {
	var details []models.Modler
	cur, err := d.collection.Find(ctx, bson.M{})
	if err != nil {
		logger.MustDebug("error occurred while fetching details")
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var detail models.Detail
		if err := cur.Decode(&detail); err != nil {
			logger.MustDebug(fmt.Sprintf("error decoding detail to json, %v", err))
			continue
		}
		details = append(details, detail)
	}

	if err := cur.Err(); err != nil {
		logger.MustDebug("error occurred with cursor")
		return nil, err
	}
	return details, nil
}
