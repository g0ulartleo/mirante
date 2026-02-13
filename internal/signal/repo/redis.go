package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	redis *redis.Client
}

func NewRedisSignalRepository() (*RedisStore, error) {
	redis := redis.NewClient(&redis.Options{
		Addr: config.Env().RedisAddr,
	})
	if err := redis.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisStore{redis: redis}, nil
}

func (r *RedisStore) Init() error {
	return nil
}

func (r *RedisStore) Close() error {
	return r.redis.Close()
}

func (r *RedisStore) Save(sig signal.Signal) error {
	ctx := context.Background()
	signalJSON, err := json.Marshal(sig)
	if err != nil {
		return err
	}
	key := "signals:" + sig.AlarmID
	if err := r.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(sig.Timestamp.Unix()),
		Member: string(signalJSON),
	}).Err(); err != nil {
		return err
	}
	// set expiry on all signals from this alarm (30 days)
	r.redis.Expire(ctx, key, 30*24*time.Hour)

	lastSignalKey := "last_signal:" + sig.AlarmID
	r.redis.Set(ctx, lastSignalKey, string(signalJSON), 30*24*time.Hour)

	return nil
}

func (r *RedisStore) GetAlarmLatestSignals(alarmID string, limit int) ([]signal.Signal, error) {
	ctx := context.Background()
	key := "signals:" + alarmID
	results, err := r.redis.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	signals := make([]signal.Signal, 0, len(results))
	for _, result := range results {
		var sig signal.Signal
		if err := json.Unmarshal([]byte(result), &sig); err != nil {
			fmt.Printf("error unmarshalling signal: %v", err)
			continue
		}
		signals = append(signals, sig)
	}

	return signals, nil
}

func (r *RedisStore) GetAlarmHealth(alarmID string) (signal.Status, error) {
	ctx := context.Background()
	lastSignalKey := "last_signal:" + alarmID

	result, err := r.redis.Get(ctx, lastSignalKey).Result()
	if err == nil {
		var sig signal.Signal
		if err := json.Unmarshal([]byte(result), &sig); err != nil {
			fmt.Printf("error unmarshalling signal: %v", err)
			return signal.StatusUnknown, nil
		}
		return sig.Status, nil
	}

	signals, err := r.GetAlarmLatestSignals(alarmID, 1)
	if err != nil || len(signals) == 0 {
		return signal.StatusUnknown, nil
	}

	return signals[0].Status, nil
}

func (r *RedisStore) CleanOldSignals() error {
	ctx := context.Background()
	iter := r.redis.Scan(ctx, 0, "signals:*", 100).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		// removing signals older than 1 week
		cutoff := time.Now().AddDate(0, 0, -7).Unix()
		_, err := r.redis.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(cutoff, 10)).Result()
		if err != nil {
			continue
		}
	}

	return iter.Err()
}

func (r *RedisStore) CountUnhealthySince(alarmID string, since time.Time) (int, error) {
	ctx := context.Background()
	key := "signals:" + alarmID
	min := fmt.Sprintf("%d", since.Unix())
	results, err := r.redis.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: min, Max: "+inf"}).Result()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, result := range results {
		var sig signal.Signal
		if err := json.Unmarshal([]byte(result), &sig); err != nil {
			continue
		}
		if sig.Status == signal.StatusUnhealthy {
			count++
		}
	}
	return count, nil
}
