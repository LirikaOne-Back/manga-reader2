package redis

import (
	"context"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/repository"
	"manga-reader2/internal/infrastructure/db"
	"time"
)

// CacheRepository реализация интерфейса repository.CacheRepository для Redis
type CacheRepository struct {
	client *db.RedisClient
	log    logger.Logger
}

// NewCacheRepository создает новый экземпляр CacheRepository
func NewCacheRepository(client *db.RedisClient, log logger.Logger) repository.CacheRepository {
	return &CacheRepository{
		client: client,
		log:    log,
	}
}

// Get получает значение по ключу
func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key)
	if err != nil {
		if err.Error() != "redis: nil" {
			r.log.Error("Ошибка получения значения из Redis", "key", key, "error", err.Error())
		}
		return "", err
	}

	return value, nil
}

// Set устанавливает значение по ключу с указанным временем жизни
func (r *CacheRepository) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration)
	if err != nil {
		r.log.Error("Ошибка установки значения в Redis", "key", key, "error", err.Error())
		return err
	}

	return nil
}

// Delete удаляет ключ
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Delete(ctx, key)
	if err != nil {
		r.log.Error("Ошибка удаления ключа из Redis", "key", key, "error", err.Error())
		return err
	}

	return nil
}

// Exists проверяет существование ключа
func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key)
	if err != nil {
		r.log.Error("Ошибка проверки существования ключа в Redis", "key", key, "error", err.Error())
		return false, err
	}

	return exists, nil
}

// Incr увеличивает значение ключа на 1
func (r *CacheRepository) Incr(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Incr(ctx, key)
	if err != nil {
		r.log.Error("Ошибка инкремента значения в Redis", "key", key, "error", err.Error())
		return 0, err
	}

	return value, nil
}

// IncrBy увеличивает значение ключа на указанное число
func (r *CacheRepository) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := r.client.IncrBy(ctx, key, value)
	if err != nil {
		r.log.Error("Ошибка инкремента значения на число в Redis", "key", key, "value", value, "error", err.Error())
		return 0, err
	}

	return result, nil
}

// ZAdd добавляет элемент в отсортированное множество
func (r *CacheRepository) ZAdd(ctx context.Context, key string, score float64, member string) error {
	err := r.client.ZAdd(ctx, key, score, member)
	if err != nil {
		r.log.Error("Ошибка добавления элемента в отсортированное множество", "key", key, "member", member, "score", score, "error", err.Error())
		return err
	}

	return nil
}

// ZIncrBy увеличивает score элемента в отсортированном множестве
func (r *CacheRepository) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	score, err := r.client.ZIncrBy(ctx, key, increment, member)
	if err != nil {
		r.log.Error("Ошибка инкремента score в отсортированном множестве", "key", key, "member", member, "increment", increment, "error", err.Error())
		return 0, err
	}

	return score, nil
}

// ZRevRange возвращает элементы из отсортированного множества в обратном порядке
func (r *CacheRepository) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	members, err := r.client.ZRevRange(ctx, key, start, stop)
	if err != nil {
		r.log.Error("Ошибка получения элементов из отсортированного множества", "key", key, "start", start, "stop", stop, "error", err.Error())
		return nil, err
	}

	return members, nil
}

// ZRevRangeWithScores возвращает элементы с их оценками из отсортированного множества в обратном порядке
func (r *CacheRepository) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) (map[string]float64, error) {
	result, err := r.client.ZRevRangeWithScores(ctx, key, start, stop)
	if err != nil {
		r.log.Error("Ошибка получения элементов со scores из отсортированного множества", "key", key, "start", start, "stop", stop, "error", err.Error())
		return nil, err
	}

	// Преобразуем результат в формат map[string]float64
	scoreMap := make(map[string]float64)
	for _, z := range result {
		if member, ok := z.Member.(string); ok {
			scoreMap[member] = z.Score
		}
	}

	return scoreMap, nil
}
