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

type SkillStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (s SkillStore) Client() *mongo.Client         { return s.client }
func (s SkillStore) Collection() *mongo.Collection { return s.collection }

func newSkillStore(ctx context.Context, uri, dbname, collectionName string) (*SkillStore, error) {
	clientOps := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOps)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error connecting to mongo client, %s", err))
		logger.MustDebug(fmt.Sprintf("error mongo db ops uri: %s", clientOps.GetURI()))
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.MustDebug("could not ping mong client")
		return nil, err
	}

	collection := client.Database(dbname).Collection(collectionName)
	return &SkillStore{client: client, collection: collection}, nil
}

func (s SkillStore) Insert(ctx context.Context, skl models.Skill) (primitive.ObjectID, error) {
	result, err := s.collection.InsertOne(ctx, skl)
	if err != nil {
		insertErr := fmt.Errorf("error inserting record, %w", err)
		logger.MustDebug("inserting skill failed")
		return primitive.NilObjectID, insertErr
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (s SkillStore) FetchById(ctx context.Context, id int) (models.Modler, error) {
	var skill models.Skill
	filter := bson.M{"_id": id}
	err := s.collection.FindOne(ctx, filter).Decode(&skill)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		} else {
			logger.MustDebug(fmt.Sprintf("error occurred while fetching skill by id, %s", err))
		}
		return nil, err
	}

	return &skill, nil
}

func (s SkillStore) FetchByName(ctx context.Context, name string) (models.Modler, error) {
	var skill models.Skill
	filter := bson.M{"name": name}
	err := s.collection.FindOne(ctx, filter).Decode(&skill)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		logger.MustDebug("error occurred while fetching skill by name")
		return nil, err
	}

	return &skill, nil
}

func (s SkillStore) Fetch(ctx context.Context) ([]models.Modler, error) {
	var skills []models.Modler
	cur, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error occurred while fetching skill, %s", err))
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var skill models.Skill
		if err := cur.Decode(&skill); err != nil {
			logger.MustDebug(fmt.Sprintf("error decoding experience to json, %v", err))
			continue
		}
		skills = append(skills, skill)
	}

	if err := cur.Err(); err != nil {
		logger.MustDebug("error occurred with cursor")
		return nil, err
	}
	return skills, nil
}
