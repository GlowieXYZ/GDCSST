package main

import (
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func connectRedis() redis.Client {
	// Connect to Redis
	RedisDB, _ := strconv.ParseInt(AppConfig.RedisDB, 10, 8)
	rdb := redis.NewClient(&redis.Options{
		Addr:     AppConfig.RedisIP + ":" + AppConfig.RedisPort,
		Password: AppConfig.RedisPassword, // no password set
		DB:       int(RedisDB),            // use default DB
	})
	return *rdb
}

func closeRedis(rdb *redis.Client) {
	rdb.Close()
}

// Helper Functions
func redisGet(key string) string {
	rdb := connectRedis()
	redisData, err := rdb.Get(ctx, key).Result()
	closeRedis(&rdb)

	if err != nil {
		panic(err)
	}

	if err == redis.Nil {
		return ""
	}
	return redisData

}

func redisSet(key string, data []byte, ttl time.Duration) {
	rdb := connectRedis()
	err := rdb.Set(ctx, key, data, ttl).Err()
	closeRedis(&rdb)
	if err != nil {
		panic(err)
	}
}

func fetchEnv(variable string, defaultVar string) string {
	val, ok := os.LookupEnv(variable)
	if ok {
		return val
	}
	return defaultVar

}

func processConfig() Config {

	AppConfig := Config{
		ListenAddress: ":9200",
		EDUser:        fetchEnv("DCS_SERVER_TRACKER_ED_USER", ""),
		EDPass:        fetchEnv("DCS_SERVER_TRACKER_ED_PASS", ""),
		RedisIP:       fetchEnv("DCS_SERVER_TRACKER_REDIS_IP", "localhost"),
		RedisPort:     fetchEnv("DCS_SERVER_TRACKER_REDIS_PORT", "6379"),
		RedisPassword: fetchEnv("DCS_SERVER_TRACKER_REDIS_PASSWORD", ""),
		RedisDB:       fetchEnv("DCS_SERVER_TRACKER_REDIS_DB", "0"),
		GeoIP2File:    fetchEnv("DCS_SERVER_TRACKER_GEOIP2_FILE", "GeoLite2-Country.mmdb"),
		PromPNGURL:    fetchEnv("DCS_SERVER_TRACKER_PROMPNG_URL", "http://localhost:8080/"),
	}
	return AppConfig
}