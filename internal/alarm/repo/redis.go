package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisAlarmRepository struct {
	redis *redis.Client
}

func NewRedisAlarmRepository() (*RedisAlarmRepository, error) {
	r := &RedisAlarmRepository{
		redis: redis.NewClient(&redis.Options{
			Addr: config.Env().RedisAddr,
		}),
	}
	if err := r.redis.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RedisAlarmRepository) Init() error {
	return nil
}

func (r *RedisAlarmRepository) GetAlarms() ([]*alarm.Alarm, error) {
	iter := r.redis.Scan(context.Background(), 0, "alarm:*", 1000).Iterator()
	alarms := make([]*alarm.Alarm, 0)
	for iter.Next(context.Background()) {
		key := iter.Val()
		alarmID := strings.TrimPrefix(key, "alarm:")
		alarm, err := r.GetAlarm(alarmID)
		if err != nil {
			return nil, err
		}
		alarms = append(alarms, alarm)
	}
	return alarms, nil
}

func (r *RedisAlarmRepository) GetAlarm(alarmID string) (*alarm.Alarm, error) {
	key := fmt.Sprintf("alarm:%s", alarmID)
	result, err := r.redis.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	var a alarm.Alarm
	if err := json.Unmarshal([]byte(result), &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *RedisAlarmRepository) SetAlarm(a *alarm.Alarm) error {
	key := fmt.Sprintf("alarm:%s", a.ID)
	alarmJSON, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if err := r.redis.Set(context.Background(), key, alarmJSON, 0).Err(); err != nil {
		return err
	}
	return r.redis.Save(context.Background()).Err()
}

func (r *RedisAlarmRepository) DeleteAlarm(alarmID string) error {
	key := fmt.Sprintf("alarm:%s", alarmID)
	if err := r.redis.Del(context.Background(), key).Err(); err != nil {
		return err
	}
	return r.redis.Save(context.Background()).Err()
}

func (r *RedisAlarmRepository) DeleteStaleAlarmsByRuntime(runtime string, keepIDs map[string]bool) error {
	all, err := r.GetAlarms()
	if err != nil {
		return err
	}
	for _, a := range all {
		if a.Runtime == runtime && !keepIDs[a.ID] {
			if err := r.DeleteAlarm(a.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RedisAlarmRepository) Close() error {
	return r.redis.Close()
}
