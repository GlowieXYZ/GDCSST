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

	for {
		t := time.Now()
		currDate := int(t.Unix())

		// Fetch the DCS servers from dcs.com
		fetchDCSServers()

		// Update the prometheus exporter
		updatePrometheus()

		processDCSServers(currDate)

		// Update the maxims
		updateMaxims(currDate)

		// Store when the Last Update was performed
		redisSet("LastUpdate", []byte(strconv.Itoa(currDate)), 0)

		// Sleep a minute
		time.Sleep(60 * time.Second)
	} // For-loop
} // Function

func fetchDCSServers() {
	log.Println("Servers: Fetching Active Server List ")

	data := url.Values{
		"AUTH_FORM":     {"Y"},
		"TYPE":          {"AUTH"},
		"backurl":       {"/en/personal/server/?ajax=y"},
		"USER_LOGIN":    {AppConfig.EDUser},
		"USER_PASSWORD": {AppConfig.EDPass},
		"USER_REMEMBER": {"Y"},
	}

	resp, err := http.PostForm("https://www.digitalcombatsimulator.com/en/personal/server/?ajax=y", data)
	if err != nil {
		log.Fatalln(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &servers)
	if err != nil {
		log.Fatalln(err)
	}
} // End Function fetchDCSServers()

func updatePrometheus() {
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
} // End Function: updatePrometheus()

// Process the DCS Servers
func processDCSServers(currDate int) {
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
	var serversProcessed int
	var serversIgnored int
	var ServerList []ServerMeta
	var metadata ServerMeta

	// Compare with Redis + Provide data for webpage
	for _, server := range servers.Servers {

		// Check if a server is blacklisted
		if strings.Contains(server.Name, AppConfig.BlackList) == true {
			serversIgnored++
			continue
		}

		// Create the unique serverID
		serverID := server.IPAddress + ":" + server.Port

		// Server Password
		var serverPassword bool

		// " Thank you for your message. Spaces are inserted intentionally after every 20 symbol to reduce table width in server and mission name. " -- ED IT Team
		// Thanks...😒
		if strings.Index(server.MissionName, " ") == 20 {
			server.MissionName = strings.ReplaceAll(server.MissionName, " ", "")
		}
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

		//		var CountryName string
		//		var CountryISO string

		//		if record.Country.Names["en"] == "China" {
		//			CountryName = "Taiwan"
		//			CountryISO = "tw"
		//		} else {
		//			CountryName = record.Country.Names["en"]
		//			CountryISO = strings.ToLower(record.Country.IsoCode)
		//		}

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

		ServerList = append(ServerList, metadata)
		serversProcessed++
	} // End of Server processing function
	log.Printf("Servers: %d processed, %d ignored.", serversProcessed, serversIgnored)

	msg, err := json.Marshal(ServerList)
	if err != nil {
		panic(err)
	}
	redisSet("metadata", msg, time.Hour)
} // End Function: proccessDCSServers()

func updateMaxims(currDate int) {
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

	maximsJSON := redisGet("maxims")
	if maximsJSON == "" {
		log.Println("Maxims not found. Initalizing.")
		localMaxims := GlobalMaxims{MaxServers: maxServers, MaxServersTime: currDate, MaxPlayers: maxPlayers, MaxPlayersTime: currDate}
		msg, errr := json.Marshal(localMaxims)
		if errr != nil {
			log.Fatal(errr)
		}
		redisSet("maxims", []byte(msg), 0)
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
			redisSet("maxims", []byte(msg), 0)
		}
	}
} // End Function: updateMaxims()
