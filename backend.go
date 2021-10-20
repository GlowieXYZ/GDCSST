package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/oschwald/geoip2-golang"
)

var servers DCSServers

func updateServers() {

	data := url.Values{
		"AUTH_FORM":     {"Y"},
		"TYPE":          {"AUTH"},
		"backurl":       {"/en/personal/server/?ajax=y"},
		"USER_LOGIN":    {AppConfig.EDUser},
		"USER_PASSWORD": {AppConfig.EDPass},
		"USER_REMEMBER": {"Y"},
	}

	for {
		t := time.Now()
		currDate := int(t.Unix())

		log.Println("Fetching server meta-data")

		resp, err := http.PostForm("https://www.digitalcombatsimulator.com/en/personal/server/?ajax=y", data)
		if err != nil {
			log.Fatalln(err)
		}

		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &servers)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Fetched server meta-data")

		// serverID Unique identifier for a server
		var serverID string

		// Clean up first
		for _, server := range servers.Servers {
			serverID = server.IPAddress + ":" + server.Port
			GdcsstPlayers.WithLabelValues(serverID).Set(float64(0))
		}

		GdcsstOnline.WithLabelValues("servers").Set(math.Round(float64(len(servers.Servers))))
		GdcsstOnline.WithLabelValues("players").Set(math.Round(float64(servers.PlayersCount)))

		for _, server := range servers.Servers {
			p, _ := strconv.ParseFloat(server.Players, 32)
			serverID = server.IPAddress + ":" + server.Port
			GdcsstPlayers.WithLabelValues(serverID).Set(p)
		}

		//log.Println("Num Servers", len(servers.Servers))

		// Connect to the GeoIP2 database
		db, err := geoip2.Open(AppConfig.GeoIP2File)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Connect to Redis
		RedisDB, _ := strconv.ParseInt(AppConfig.RedisDB, 10, 8)
		rdb := redis.NewClient(&redis.Options{
			Addr:     AppConfig.RedisIP + ":" + AppConfig.RedisPort,
			Password: AppConfig.RedisPassword,
			DB:       int(RedisDB),
		})
		defer db.Close()

		// Required for the next step
		cacheIgnored := 0
		cacheResolved := 0

		var ServerList []ServerMeta
		var metadata ServerMeta

		// Compare with Redis + Provide data for webpage
		for _, server := range servers.Servers {
			// Create the unique serverID
			serverID = server.IPAddress + ":" + server.Port

			// Server Password
			var serverPassword bool

			// Server stuff
			serverMissionTime, _ := strconv.ParseInt(server.MissionTime, 10, 8)
			serverPlayers, _ := strconv.ParseInt(server.Players, 10, 8)
			serverPlayersMax, _ := strconv.ParseInt(server.PlayersMax, 10, 8)

			if server.Password == "Yes" {
				serverPassword = true
			} else {
				serverPassword = false
			}

			// Do IP stuff
			ip := net.ParseIP(server.IPAddress)
			record, err := db.Country(ip)
			if err != nil {
				log.Fatal(err)
			}

			// Get record from Redis
			redisServer, err := rdb.Get(ctx, serverID).Result()
			if err == redis.Nil {

				metadata = ServerMeta{
					Name:                 server.Name,
					IPAddress:            server.IPAddress,
					Port:                 server.Port,
					MissionName:          server.MissionName,
					MissionTime:          int(serverMissionTime),
					Players:              int(serverPlayers),
					PlayersMax:           int(serverPlayersMax),
					Password:             serverPassword,
					Description:          server.Description,
					MissionTimeFormatted: server.MissionTimeFormatted,
					CountryName:          record.Country.Names["en"],
					ISO:                  strings.ToLower(record.Country.IsoCode),
					LastUpdate:           currDate,
				}

				msg, errr := json.Marshal(metadata)
				if errr != nil {
					log.Fatal(errr)
				}

				err = rdb.Set(ctx, serverID, msg, 750*time.Hour).Err()
				if err != nil {
					panic(err)
				}

				cacheResolved++
			} else if err != nil {
				panic(err)
			} else {
				redisServerB := []byte(redisServer)
				json.Unmarshal(redisServerB, &metadata)

				metadata = ServerMeta{
					Name:                 server.Name,
					IPAddress:            server.IPAddress,
					Port:                 server.Port,
					MissionName:          server.MissionName,
					MissionTime:          int(serverMissionTime),
					Players:              int(serverPlayers),
					PlayersMax:           int(serverPlayersMax),
					Password:             serverPassword,
					Description:          server.Description,
					MissionTimeFormatted: server.MissionTimeFormatted,
					CountryName:          record.Country.Names["en"],
					ISO:                  strings.ToLower(record.Country.IsoCode),
					LastUpdate:           currDate,
				}

				// Write stuff
				msg, errr := json.Marshal(metadata)
				if errr != nil {
					log.Fatal(errr)
				}

				err = rdb.Set(ctx, serverID, msg, 750*time.Hour).Err()
				if err != nil {
					panic(err)
				}

				cacheIgnored++
			}
			if strings.Contains(server.Name, AppConfig.BlackList) == false {
				ServerList = append(ServerList, metadata)
			}
		} // End of Server processing function
		log.Printf("Done processing servers: %d processed and %d ignored.", cacheResolved, cacheIgnored)

		msg, err := json.Marshal(ServerList)
		if err != nil {
			panic(err)
		}
		redisSet("metadata", msg, time.Hour)

		// Handle max:
		var maxPlayers int
		var maxServers int
		var maxUpdate int

		for _, server := range servers.Servers {
			players, _ := strconv.ParseInt(server.Players, 10, 8)
			maxPlayers += int(players)
			maxServers++
		}

		var localMaxims GlobalMaxims

		maximsJSON, err := rdb.Get(ctx, "maxims").Result()
		if err == redis.Nil {
			log.Println("Maxims not found. Initalizing.")
			localMaxims := GlobalMaxims{MaxServers: maxServers, MaxServersTime: currDate, MaxPlayers: maxPlayers, MaxPlayersTime: currDate}

			msg, errr := json.Marshal(localMaxims)
			if errr != nil {
				log.Fatal(errr)
			}

			err = rdb.Set(ctx, "maxims", msg, 0).Err()
			if err != nil {
				panic(err)
			}

		} else if err != nil {
			panic(err)
		} else {
			maximsB := []byte(maximsJSON)
			json.Unmarshal(maximsB, &localMaxims)

			if localMaxims.MaxServers < maxServers {
				log.Printf("Servers %d > %d", maxServers, localMaxims.MaxServers)
				localMaxims.MaxServers = maxServers
				localMaxims.MaxServersTime = currDate
				maxUpdate = 1
			}

			if localMaxims.MaxPlayers < maxPlayers {
				log.Printf("Players %d > %d", maxPlayers, localMaxims.MaxPlayers)
				localMaxims.MaxPlayers = maxPlayers
				localMaxims.MaxPlayersTime = currDate
				maxUpdate = 1
			}

			if maxUpdate == 1 {
				log.Println("New Maxims found, updating! \\o/")
				msg, errr := json.Marshal(localMaxims)
				if errr != nil {
					log.Fatal(errr)
				}

				err = rdb.Set(ctx, "maxims", msg, 0).Err()
				if err != nil {
					panic(err)
				}
			}
		}

		redisSet("LastUpdate", []byte(strconv.Itoa(currDate)), 0)

		// Sleep a minute
		time.Sleep(60 * time.Second)
	} // For-loop
} // Function
