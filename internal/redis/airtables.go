package redis

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var QueueKey = os.Getenv("TRAIL_NAME") + ":queued_tables"
var PointerKey = os.Getenv("TRAIL_NAME") + ":queue_pointer"

var TablesInitialized bool = false

var advancePointerScript = redis.NewScript(`
	local tableName = KEYS[1]
	local pointer = ARGV[1]
    local current = redis.call("GET", tableName)
    if current == pointer then
        redis.call("SET", KEYS[1], pointer + 1)
        return 1
    end
    return 0
`)

func reshuffleQueue() error {
	ctx := context.Background()

	tables, err := client.LRange(ctx, QueueKey, 0, -1).Result()
	if err != nil {
		return err
	}

	rand.Shuffle(len(tables), func(i, j int) {
		tables[i], tables[j] = tables[j], tables[i]
	})

	_, err = client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, QueueKey)
		pipe.RPush(ctx, QueueKey, tables)
		pipe.Set(ctx, PointerKey, 0, 0)
		return nil
	})
	return err
}

func InitAirtableSets(tables []string) error {
	if client == nil {
		return fmt.Errorf("Redis has not been initalized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	set, err := client.SetNX(ctx, QueueKey+":lock", 1, 0).Result()
	if err != nil {
		return err
	}
	if !set {
		return nil
	}

	rand.Shuffle(len(tables), func(i, j int) {
		tables[i], tables[j] = tables[j], tables[i]
	})

	_, err = client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.RPush(ctx, QueueKey, tables)
		pipe.Set(ctx, PointerKey, 0, 0)
		return nil
	})

	TablesInitialized = true
	return err
}

func GetNextTable() (string, error) {
	if !TablesInitialized {
		return "", fmt.Errorf("Airtable list is not initalized in redis cache!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		length, err := client.LLen(ctx, QueueKey).Result()
		if err != nil {
			return "", err
		}

		pointer, err := client.Get(ctx, PointerKey).Int64()
		if err != nil {
			return "", err
		}

		if pointer >= length {
			if err := reshuffleQueue(); err != nil {
				return "", err
			}
			pointer = 0
		}

		table, err := client.LIndex(ctx, QueueKey, pointer).Result()
		if err != nil {
			return "", err
		}

		swapped, err := advancePointerScript.Run(ctx, client,
			[]string{PointerKey},
			strconv.FormatInt(pointer, 10),
		).Int()
		if err != nil {
			return "", err
		}

		if swapped == 1 {
			return table, nil
		}
	}
}
