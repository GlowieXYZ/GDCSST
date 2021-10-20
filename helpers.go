package main

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Println("Redis error kicked in:", err, " This was with key", key)
		return ""
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

func redisKeys(search string) []string {
	rdb := connectRedis()
	val, err := rdb.Keys(ctx, search).Result()
	closeRedis(&rdb)
	if err != nil {
		log.Println("Redis key() failed:", err)
	}
	return val
}

func fetchEnv(variable string, defaultVar string) string {
	val, ok := os.LookupEnv(variable)
	if ok {
		return val
	}
	return defaultVar

}

func findRelatedServers(server string, port string) []string {
	// Variable declarations
	var serverID string
	var relatedServers []string
	var related []string

	serverID = server + ":" + port
	related = redisKeys(server + "*")
	if len(related) > 1 {
		for _, relatedServer := range related {
			if relatedServer != serverID {
				relatedServers = append(relatedServers, relatedServer)
			}
		}
	}
	return relatedServers
} // End Function: updateRelatedServers()

func fetchServerInfo(server string) ServerMeta {
	ServerInfo := redisGet(server)
	var ServerDataMeta ServerMeta
	json.Unmarshal([]byte(ServerInfo), &ServerDataMeta)
	return ServerDataMeta
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
		BlackList:     fetchEnv("DCS_SERVER_TRACKER_BLACKLIST", ""),
		PromPNGURL:    fetchEnv("DCS_SERVER_TRACKER_PROMPNG_URL", "http://localhost:8080/"),
	}

	log.Println("Using the following config:")
	fmt.Println("DCS_SERVER_TRACKER_ED_USER:", AppConfig.EDUser)
	fmt.Println("DCS_SERVER_TRACKER_ED_PASS: ****")
	fmt.Println("DCS_SERVER_TRACKER_REDIS_IP", AppConfig.RedisIP)
	fmt.Println("DCS_SERVER_TRACKER_REDIS_PORT", AppConfig.RedisPort)
	fmt.Println("DCS_SERVER_TRACKER_REDIS_PASSWORD: ****")
	fmt.Println("DCS_SERVER_TRACKER_REDIS_DB", AppConfig.RedisDB)
	fmt.Println("DCS_SERVER_TRACKER_GEOIP2_FILE", AppConfig.GeoIP2File)
	fmt.Println("DCS_SERVER_TRACKER_BLACKLIST:", AppConfig.BlackList)

	return AppConfig
}
