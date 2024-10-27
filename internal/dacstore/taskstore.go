package dacstore

import (
	"context"
	"fmt"
	"time"

	"github.com/Z3DRP/zportfolio-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (t TaskStore) Client() *mongo.Client         { return t.client }
func (t TaskStore) Collection() *mongo.Collection { return t.collection }

func newTaskStore(client *mongo.Client, dbname, collectionName string) *TaskStore {
	collection := client.Database(dbname).Collection(collectionName)
	return &TaskStore{client: client, collection: collection}
}

func (t TaskStore) Insert(ctx context.Context, tsk models.Task) (primitive.ObjectID, error) {
	result, err := t.collection.InsertOne(ctx, tsk)
	if err != nil {
		insertErr := fmt.Errorf("error inserting Task record, %w", err)
		logger.MustDebug(fmt.Sprintf("inserting Task record failed, %v", err))
		return primitive.NilObjectID, insertErr
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (t TaskStore) FetchTaskInPeriod(ctx context.Context, start, end time.Time) (models.Tasklist, error) {
	var tasks models.Tasklist
	filter := bson.M{"$and": bson.A{
		bson.M{"start_date": bson.M{"$gte": start}},
		bson.M{"end_date": bson.M{"lte": end}},
	}}

	cur, err := t.collection.Find(ctx, filter)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error occurred while fetching tasks:: %s", err))
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var task models.Task
		if err := cur.Decode(&task); err != nil {
			logger.MustDebug(fmt.Sprintf("error decoding experience to json:: %v", err))
			continue
		}
		tasks = append(tasks, task)
	}

	if err := cur.Err(); err != nil {
		logger.MustDebug(fmt.Sprintf("error occurred with task cursor:: %v", err))
		return nil, err
	}
	return tasks, nil
}

func (t TaskStore) FetchTask(ctx context.Context, tid string) (models.Modler, error) {
	var task models.Task
	filter := bson.M{"tid": tid}
	err := t.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NewNoResultErr(tid, "skill", err)
		}
		return nil, err
	}
	return &task, nil
}

func (t TaskStore) UpdateTask(ctx context.Context, tid string, tsk *models.Task) (int64, int64, error) {
	filter := bson.M{"tid": tid}
	update := bson.D{{
		Key: "$set",
		Value: bson.M{
			"start_time": tsk.StartTime,
			"end_time":   tsk.EndTime,
			"detail":     tsk.Detail,
		},
	}}

	result, err := t.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, 0, err
	}

	return result.MatchedCount, result.ModifiedCount, nil
}

func (t TaskStore) DeleteTask(ctx context.Context, tid string) (int64, error) {
	filter := bson.M{"tid": tid}
	result, err := t.collection.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
