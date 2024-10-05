package dacstore

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/config"
	zlg "github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Store interface {
	Client() *mongo.Client
	Collection() *mongo.Collection
	Insert(interface{}) (primitive.ObjectID, error)
	FetchById(id int) (interface{}, error)
	FetchByName(name string) (interface{}, error)
	Fetch() []interface{}
}

var logfile = zlg.NewLogFile(
	zlg.WithFilename(fmt.Sprintf("%s/db/%s", config.LogPrefix, config.LogName)),
)

var logger = config.NewLogger(logfile, "trace", true, false)

var dbConfig, cfgErr = config.ReadDbConfig()

func CreateDetailStore(ctx context.Context) (*DetailStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped detail store creation, %s", cfgErr))
		return nil, fmt.Errorf("config error stopped detail store creation, %w", cfgErr)
	}
	return newDetailStore(ctx, dbConfig.DbUri, dbConfig.DbName, dbConfig.DetailCol)
}

func CreateExperienceStore(ctx context.Context) (*ExperienceStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped experience store creation, %s", cfgErr))
		return nil, fmt.Errorf("config error stopped experience store creation, %w", cfgErr)
	}
	return newExperienceStore(ctx, dbConfig.DbUri, dbConfig.DbName, dbConfig.ExpCol)
}

func CreateSkillStore(ctx context.Context) (*SkillStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped skill store creation, %s", cfgErr))
		return nil, fmt.Errorf("config error stopped skill creation")
	}
	return newSkillStore(ctx, dbConfig.DbUri, dbConfig.DbName, dbConfig.SkillCol)
}

var ErrFetchSkill = fmt.Errorf("an error occured while fetching skills")
var ErrFetchExperience = fmt.Errorf("an error occured while fetching experience")
var ErrFetchDetails = fmt.Errorf("an error occured while fetching details")
