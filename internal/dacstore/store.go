package dacstore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Z3DRP/zportfolio-service/config"
	zlg "github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ErrFailedInsert struct {
	Stype   string
	Details string
	Err     error
}

func (fi *ErrFailedInsert) Error() string {
	return fmt.Sprintf("failed to insert record: %v{%v}:: %v", fi.Stype, fi.Details, fi.Err)
}

func (fi *ErrFailedInsert) Unwrap() error {
	return fi.Err
}

type ErrNoResults struct {
	Identifier string
	Object     string
	Err        error
}

func (e *ErrNoResults) Error() string {
	return fmt.Sprintf("no results found for Type: %v, with identifier: %v", e.Object, e.Identifier)
}

func (e *ErrNoResults) Unwrap() error {
	return e.Err
}

func NewNoResultErr(identifer, obj string, err error) *ErrNoResults {
	return &ErrNoResults{
		Identifier: identifer,
		Object:     obj,
		Err:        err,
	}
}

var logfile = zlg.NewLogFile(
	zlg.WithFilename(fmt.Sprintf("%v/%v", config.LogPrefix, "data.log")),
)

var logger = zlg.NewLogger(
	logfile,
	zlg.WithJsonFormatter(true),
	zlg.WithLevel("trace"),
	zlg.WithReportCaller(false),
)
var dbConfig, cfgErr = config.ReadDbConfig()
var mClient *mongo.Client
var clientInitError error
var mongoClientOnce sync.Once

var ErrFetchSkill = fmt.Errorf("an error occured while fetching skills")
var ErrFetchExperience = fmt.Errorf("an error occured while fetching experience")
var ErrFetchDetails = fmt.Errorf("an error occured while fetching details")
var ErrFetchTask = fmt.Errorf("an error occurred while fetching tasks")
var ErrFetchAvailability = fmt.Errorf("an error occurred while fetching availability")
var errMongoConn = fmt.Errorf("could not create connection to mongo cluster")

// TODO add an storeErr file to package and move errors there and make error for config err
type Store interface {
	Client() *mongo.Client
	Collection() *mongo.Collection
	Insert(interface{}) (primitive.ObjectID, error)
	FetchById(id int) (interface{}, error)
	FetchByName(name string) (interface{}, error)
	Fetch() []interface{}
}

func CreateDetailStore(ctx context.Context) (*DetailStore, error) {
	if cfgErr != nil {
		logger.MustTrace(fmt.Sprintf("config error stopped detail store creation, %s", cfgErr))
		return nil, fmt.Errorf("config error stopped detail store creation, %w", cfgErr)
	}

	logger.MustTrace("initializing mongo conn")
	mc, err := initializeMongoConnection(ctx, dbConfig.DbUri)
	logger.MustTrace(fmt.Sprintf("connection created:: %v", mc))

	if err != nil {
		logger.MustTrace(fmt.Sprintf("error durring init con:: %s", err))
		return nil, fmt.Errorf("error initializing mongo connection: %w", err)
	}

	return newDetailStore(mc, dbConfig.DbName, dbConfig.DetailCol), nil
}

func CreateExperienceStore(ctx context.Context) (*ExperienceStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped experience store creation, %s", cfgErr))
		return nil, fmt.Errorf("config error stopped experience store creation, %w", cfgErr)
	}

	mc, err := initializeMongoConnection(ctx, dbConfig.DbUri)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", errMongoConn, err)
	}
	return newExperienceStore(mc, dbConfig.DbName, dbConfig.ExpCol), nil
}

func CreateSkillStore(ctx context.Context) (*SkillStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped skill store creation, %s", cfgErr))
		return nil, fmt.Errorf("config error stopped skill creation")
	}

	mc, err := initializeMongoConnection(ctx, dbConfig.DbUri)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMongoConn, err)
	}
	return newSkillStore(mc, dbConfig.DbName, dbConfig.SkillCol), nil
}

func CreateVisitorStore(ctx context.Context) (*VisitorStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped visitor store creation:: %v", cfgErr))
		return nil, fmt.Errorf("config error stopped visitor store creation:: %w", cfgErr)
	}

	mc, err := initializeMongoConnection(ctx, dbConfig.DbUri)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMongoConn, err)
	}
	return newVisitorStore(mc, dbConfig.DbName, dbConfig.VisitorCol), nil
}

func CreateTaskStore(ctx context.Context) (*TaskStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped task store creation:: %v", cfgErr))
		return nil, fmt.Errorf("config error stopped task store creation")
	}

	mc, err := initializeMongoConnection(ctx, dbConfig.DbUri)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMongoConn, err)
	}
	return newTaskStore(mc, dbConfig.DbName, dbConfig.TaskCol), nil
}

func CreateAvailabilityStore(ctx context.Context) (*AvailabilityStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped availability store creation:: %v", cfgErr))
		return nil, fmt.Errorf("config error stopped availability store creation")
	}

	mc, err := initializeMongoConnection(ctx, dbConfig.DbUri)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMongoConn, err)
	}
	return newAvailabilityStore(mc, dbConfig.DbName, dbConfig.AvailabilityCol), nil
}

func CreatePeriodStore(ctx context.Context) (*PeriodStore, error) {
	if cfgErr != nil {
		logger.MustDebug(fmt.Sprintf("config error stopped period store creation, %s", cfgErr))
		return nil, fmt.Errorf("config error stopped period store creation")
	}

	mc, err := initializeMongoConnection(ctx, dbConfig.DbUri)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMongoConn, err)
	}
	return newPeriodStore(mc, dbConfig.DbName, dbConfig.PeriodCol), nil
}

func initializeMongoConnection(ctx context.Context, uri string) (*mongo.Client, error) {
	mongoClientOnce.Do(func() {
		clientOps := options.Client().ApplyURI(uri).
			SetMaxPoolSize(120).SetMinPoolSize(30).
			SetMaxConnecting(50).
			SetMaxConnIdleTime(5 * time.Minute)

		mClient, clientInitError = mongo.Connect(ctx, clientOps)
		if clientInitError != nil {
			logger.MustDebug(fmt.Sprintf("an error occurred while creating connection to mongo: %s", clientInitError))
		}

		if clientInitError = mClient.Ping(ctx, nil); clientInitError != nil {
			logger.MustDebug(fmt.Sprintf("failed to ping mongo db cluster: %s", clientInitError))
		}
	})

	return mClient, clientInitError
}
