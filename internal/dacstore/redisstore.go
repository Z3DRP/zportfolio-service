package dacstore

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/redis/go-redis/v9"
)

var rClient *redis.Client
var redisOnce sync.Once
var redisConErr error

type ErrNoCacheResult struct {
	ClientId *redis.IntCmd
	Key      string
	Err      error
}

func (e *ErrNoCacheResult) Error() string {
	return fmt.Sprintf("cacheId: %v; does not have a value for Key: %v", e.ClientId, e.Key)
}

func (e *ErrNoCacheResult) Unwrap() error {
	return e.Err
}

func NewNoCacheResultErr(cid *redis.IntCmd, k string, err error) *ErrNoCacheResult {
	return &ErrNoCacheResult{
		ClientId: cid,
		Key:      k,
		Err:      err,
	}
}

func NewRedisClient(ctx context.Context) (*redis.Client, error) {
	redisOnce.Do(func() {
		rClient = redis.NewClient(&redis.Options{
			Addr:     "redis-service:6379",
			Password: "",
			DB:       0,
		})
		_, err := rClient.Ping(ctx).Result()
		if err != nil {
			rClient = nil
			redisConErr = err
		}
	})
	logger.MustTrace(fmt.Sprintf("returning redis client: %v", rClient))
	return rClient, redisConErr
}

func SetZypherValue(ctx context.Context, client *redis.Client, txt string) error {
	err := client.Set(ctx, txt, txt, 0).Err()
	return err
}

func CheckZypherValue(ctx context.Context, client *redis.Client, txt string) (string, error) {
	val, err := client.Get(ctx, txt).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func SetPortfolioData(ctx context.Context, client *redis.Client, data models.Responser) error {
	jsonData, err := json.Marshal(data)
	logger.MustDebug(fmt.Sprintf("setting cache with:: client: %v; data: %s", client, jsonData))
	if err != nil {
		return err
	}
	err = client.Set(ctx, "ZachPalmer", jsonData, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func CheckPortfolioData(ctx context.Context, client *redis.Client) (models.Responser, error) {

	val, err := client.Get(ctx, "ZachPalmer").Result()
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error occurred while getting portfolio cache:: %v", err))
		if err != redis.Nil {
			logger.MustDebug(fmt.Sprintf("unexpected cache error:: %v", err))
			return nil, err
		}
		return nil, NewNoCacheResultErr(client.ClientID(ctx), "Zach Palmer", err)
	}

	// if val == "" {
	// 	logger.MustTrace("initial key value.. or key was empty")
	// 	return nil, NewNoCacheResultErr(client.ClientID(ctx), "Zach Palmer", nil)
	// }
	logger.MustDebug(fmt.Sprintf("no error occurred fetching cache:: value: %v; begining to unmarshal...", val))

	var data models.PortfolioResponse
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error unmarchalling portfolio data: %v", err))
		return nil, err
	}
	return &data, nil
}

func SetScheduleData(ctx context.Context, client *redis.Client, pstart, pend time.Time, data models.Schedule) error {
	key := fmt.Sprintf("%v-%v", pstart.String(), pend.String())
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = client.Set(ctx, key, jsonData, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetScheduleData(ctx context.Context, client *redis.Client, pstart, pend time.Time) (models.Responser, error) {
	key := fmt.Sprintf("%v-%v", pstart.String(), pend.String())
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error occurred while getting schedule cache:: %v", err))
		if err != redis.Nil {
			logger.MustDebug(fmt.Sprintf("unexpected cache error:: %v", err))
			return nil, err
		}
		return nil, NewNoCacheResultErr(client.ClientID(ctx), key, err)
	}

	var data models.ScheduleResponse
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error unmarshalling schedule data: %v", err))
		return nil, err
	}
	return &data, nil
}
