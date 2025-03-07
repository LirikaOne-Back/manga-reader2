package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"manga-reader2/internal/common/logger"
)

// RedisConfig содержит настройки подключения к Redis
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// RedisClient представляет подключение к Redis
type RedisClient struct {
	client *redis.Client
	log    logger.Logger
}

// NewRedisClient создает и настраивает новое подключение к Redis
func NewRedisClient(ctx context.Context, cfg RedisConfig, log logger.Logger) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	log.Info("Подключение к Redis", "addr", addr, "db", cfg.DB)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	log.Info("Успешное подключение к Redis")

	return &RedisClient{
		client: client,
		log:    log,
	}, nil
}

// GetClient возвращает клиент Redis
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Close закрывает соединение с Redis
func (r *RedisClient) Close() error {
	r.log.Info("Закрытие соединения с Redis")
	return r.client.Close()
}

// Методы для работы с Redis

// Get получает значение по ключу
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set устанавливает значение по ключу с опциональным временем жизни
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Delete удаляет ключ(и)
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists проверяет существование ключа
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Incr увеличивает значение ключа на 1
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy увеличивает значение ключа на указанное число
func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// Expire устанавливает время жизни ключа
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, expiration).Result()
}

// ZAdd добавляет элемент в отсортированное множество
func (r *RedisClient) ZAdd(ctx context.Context, key string, score float64, member string) error {
	z := redis.Z{
		Score:  score,
		Member: member,
	}
	return r.client.ZAdd(ctx, key, z).Err()
}

// ZIncrBy увеличивает score элемента в отсортированном множестве
func (r *RedisClient) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	return r.client.ZIncrBy(ctx, key, increment, member).Result()
}

// ZRevRange возвращает элементы из отсортированного множества в обратном порядке
func (r *RedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRevRangeWithScores возвращает элементы с их оценками из отсортированного множества в обратном порядке
func (r *RedisClient) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

// Scan сканирует ключи по шаблону
func (r *RedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, match, count).Result()
}

// FlushDB очищает текущую БД Redis
func (r *RedisClient) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Subscribe подписывается на канал
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// Publish публикует сообщение в канал
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}
