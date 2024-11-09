package dacstore

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/utils"
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
	return fmt.Sprintf("cacheId: %v; does not have a value for Key: %v", e.ClientId.String(), e.Key)
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

type ErrRedisConnect struct {
	ClientId *redis.IntCmd
	Err      error
}

func (e *ErrRedisConnect) Error() string {
	return fmt.Sprintf("could not connect to redis client: %v :: %v", e.ClientId.String(), e.Err)
}

func (e *ErrRedisConnect) Unwrap() error {
	return e.Err
}

func NewRedisConnErr(cid *redis.IntCmd, e error) *ErrRedisConnect {
	return &ErrRedisConnect{
		ClientId: cid,
		Err:      e,
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

func SetScheduleData(ctx context.Context, client *redis.Client, pstart, pend time.Time, data models.Responser) error {
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

func CheckScheduleData(ctx context.Context, client *redis.Client, pstart, pend time.Time) (models.Responser, error) {
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

func SetUserData(ctx context.Context, client *redis.Client, name, company, phone, eml, roles, ip, k string) error {
	usr := adapters.NewUserData(name, company, eml, phone, roles)
	userDTO := dtos.NewUserDto(usr, ip, k)
	data, err := json.Marshal(userDTO)

	if err != nil {
		return utils.NewJsonEncodeErr(usr, err)
	}

	err = client.Set(ctx, k, data, 0).Err()
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error caching user data:: %v", err))
		return fmt.Errorf("error caching user data:: %w", err)
	}
	return nil
}

func CheckUserData(ctx context.Context, client *redis.Client, k string) (dtos.DTOer, error) {
	val, err := client.Get(ctx, k).Result()
	if err != nil {
		if err != redis.Nil {
			logger.MustDebug(fmt.Sprintf("unexpected user cache error:: %v", err))
			return nil, fmt.Errorf("unexpected user cache error:: %w", err)
		}
		return nil, NewNoCacheResultErr(client.ClientID(ctx), k, err)
	}

	var data dtos.UserDto
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error unmarshalling user data: %v", err))
		return nil, fmt.Errorf("unexpected user cache error:: %w", err)
	}
	return &data, nil
}
