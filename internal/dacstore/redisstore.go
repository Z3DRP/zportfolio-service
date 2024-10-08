package dacstore

import (
	"context"
	"encoding/json"

	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/go-redis/redis"
)

func NewRedisClient(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func SetZypherValue(txt string, client *redis.Client) error {
	err := client.Set(txt, txt, 0).Err()
	return err
}

func CheckZypherValue(txt string, client *redis.Client) (string, error) {
	val, err := client.Get(txt).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func SetPortfolioData(data models.Responser, client *redis.Client) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = client.Set("Zach Palmer", jsonData, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func CheckPortfolioData(client *redis.Client) (models.Responser, error) {
	val, err := client.Get("Zach Palmer").Result()
	if err != nil {
		return models.PortoflioResponse{}, err
	}
	var data models.PortoflioResponse
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return models.PortoflioResponse{}, err
	}
	return data, nil
}
