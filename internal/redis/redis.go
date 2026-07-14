package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/WeatherGod3218/mlc-project-template/internal/logging"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

var client *redis.Client

func InitRedis() error {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return err
	}

	newClient := redis.NewClient(opt)
	pong, err := newClient.Ping(ctx).Result()
	if err != nil {
		return err
	}

	client = newClient
	logging.Logger.WithFields(logrus.Fields{"module": "redis", "method": "InitRedis"}).Info(fmt.Sprintf("Pinged Redis!: %s", pong))
	return nil
}
