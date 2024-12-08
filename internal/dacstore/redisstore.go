package dacstore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/utils"
	zlg "github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"github.com/redis/go-redis/v9"
)

var rClient *redis.Client
var redisOnce sync.Once
var redisConErr error

var zlogfile = zlg.NewLogFile(
	zlg.WithFilename(fmt.Sprintf("%v/%v", config.LogPrefix, "cache.log")),
)

var loggr = zlg.NewLogger(
	zlogfile,
	zlg.WithJsonFormatter(true),
	zlg.WithLevel("debug"),
	zlg.WithReportCaller(false),
)

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
	loggr.MustDebug("in new client func")

	//Addr:     "127.0.0.1:6379", // development redis container
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	if host == "" {
		host = "redis-service"
	}
	if port == "" {
		port = "6379"
	}
	adr := fmt.Sprintf("%v:%v", host, port)

	redisOnce.Do(func() {
		rClient = redis.NewClient(&redis.Options{
			Addr: "redis-service:6379", //original prod redis container change back
			//Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		})
		pong, err := rClient.Ping(ctx).Result()

		loggr.MustDebug(fmt.Sprintf("result from ping in go container: %v", pong))

		if err != nil {
			rClient = nil
			redisConErr = err
		}
	})

	if rClient == nil {
		initializeRedisCon(ctx, adr)
	}

	pong, perr := rClient.Ping(ctx).Result()
	if perr != nil {
		logger.MustDebug(fmt.Sprintf("the following error occurred when pinging redis: %v", perr))
	}

	if pong != "" {
		logger.MustDebug(fmt.Sprintf("the redis pong result: %v", pong))
	}

	loggr.MustDebug(fmt.Sprintf("returning redis client: %v, redis address: %v", rClient, adr))

	return rClient, redisConErr
}

func initializeRedisCon(ctx context.Context, adr string) {
	rClient = redis.NewClient(&redis.Options{
		//Addr: "redis-service:6379", //original prod redis container change back
		Addr: adr,
		//Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	pong, err := rClient.Ping(ctx).Result()

	loggr.MustDebug(fmt.Sprintf("result from ping in go container: %v", pong))

	if err != nil {
		rClient = nil
		redisConErr = err
	}

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
	loggr.MustDebug("setting portfolio data")
	jsonData, err := json.Marshal(data)
	loggr.MustDebug(fmt.Sprintf("setting cache with:: client: %v; data: %s", client, jsonData))
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
	loggr.MustDebug("checking data....")
	val, err := client.Get(ctx, "ZachPalmer").Result()
	if err != nil {
		loggr.MustDebug(fmt.Sprintf("error occurred while getting portfolio cache:: %v", err))
		if err != redis.Nil {
			loggr.MustDebug(fmt.Sprintf("unexpected cache error:: %v", err))
			return nil, err
		}
		return nil, NewNoCacheResultErr(client.ClientID(ctx), "Zach Palmer", err)
	}

	// if val == "" {
	// 	loggr.MustTrace("initial key value.. or key was empty")
	// 	return nil, NewNoCacheResultErr(client.ClientID(ctx), "Zach Palmer", nil)
	// }
	loggr.MustDebug(fmt.Sprintf("no error occurred fetching cache:: value: %v; begining to unmarshal...", val))

	var data models.PortfolioResponse
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		loggr.MustDebug(fmt.Sprintf("error unmarchalling portfolio data: %v", err))
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
		loggr.MustDebug(fmt.Sprintf("error occurred while getting schedule cache:: %v", err))
		if err != redis.Nil {
			loggr.MustDebug(fmt.Sprintf("unexpected cache error:: %v", err))
			return nil, err
		}
		return nil, NewNoCacheResultErr(client.ClientID(ctx), key, err)
	}

	var data models.ScheduleResponse
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		loggr.MustDebug(fmt.Sprintf("error unmarshalling schedule data: %v", err))
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
		loggr.MustDebug(fmt.Sprintf("error caching user data:: %v", err))
		return fmt.Errorf("error caching user data:: %w", err)
	}
	return nil
}

func CheckUserData(ctx context.Context, client *redis.Client, k string) (dtos.DTOer, error) {
	loggr.MustDebug("checking usr data...")
	val, err := client.Get(ctx, k).Result()
	if err != nil {
		if err != redis.Nil {
			loggr.MustDebug(fmt.Sprintf("unexpected user cache error:: %v", err))
			return nil, fmt.Errorf("unexpected user cache error:: %w", err)
		}
		return nil, NewNoCacheResultErr(client.ClientID(ctx), k, err)
	}

	var data dtos.UserDto
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		loggr.MustDebug(fmt.Sprintf("error unmarshalling user data: %v", err))
		return nil, fmt.Errorf("unexpected user cache error:: %w", err)
	}
	return &data, nil
}
