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

type ExperienceStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (e ExperienceStore) Client() *mongo.Client         { return e.client }
func (e ExperienceStore) Collection() *mongo.Collection { return e.collection }

func newExperienceStore(ctx context.Context, uri, dbName, collectionName string) (*ExperienceStore, error) {
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.MustDebug("error connecting to mongo client")
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.MustDebug("could not ping mongo client")
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)
	return &ExperienceStore{client: client, collection: collection}, nil
}

func (e *ExperienceStore) Insert(ctx context.Context, xp models.Experience) (primitive.ObjectID, error) {
	result, err := e.collection.InsertOne(ctx, xp)
	if err != nil {
		insertErr := fmt.Errorf("error inserting record, %w", err)
		logger.MustDebug("inserting experience failed")
		return primitive.NilObjectID, insertErr
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (e *ExperienceStore) FetchById(ctx context.Context, id int) (models.Modler, error) {
	var xp models.Experience
	filter := bson.M{"_id": id}
	err := e.collection.FindOne(ctx, filter).Decode(&xp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		} else {
			logger.MustDebug("error occurred while fetching experience by id")
		}
		return nil, err
	}

	return &xp, nil
}

func (e *ExperienceStore) FetchByName(ctx context.Context, name string) (models.Modler, error) {
	var xp models.Experience
	filter := bson.M{"name": name}
	err := e.collection.FindOne(ctx, filter).Decode(&xp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		logger.MustDebug("error occurred while fetching by name experience")
		return nil, err
	}

	return &xp, nil
}

func (e *ExperienceStore) Fetch(ctx context.Context) ([]models.Modler, error) {
	var xps []models.Modler
	cur, err := e.collection.Find(ctx, bson.M{})
	if err != nil {
		logger.MustDebug("error occurred while fetching experience")
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var xp models.Experience
		if err := cur.Decode(&xp); err != nil {
			logger.MustDebug(fmt.Sprintf("error decoding experience to json, %v", err))
			continue
		}
		xps = append(xps, xp)
	}

	if err := cur.Err(); err != nil {
		logger.MustDebug("error occurred with cursor")
		return nil, err
	}
	return xps, nil
}
