package main

// Config This contains all the application specific config
type Config struct {
	ListenAddress string
	EDUser        string
	EDPass        string
	RedisIP       string
	RedisPort     string
	RedisPassword string
	RedisDB       string
	GeoIP2File    string
	PromPNGURL    string
	BlackList     string
}

// DCSServers is struct to capture the JSON data returning from DCS.com
type DCSServers struct {
	ServersMaxCount int           `json:"SERVERS_MAX_COUNT"`
	ServersMaxDate  string        `json:"SERVERS_MAX_DATE"`
	PlayersCount    int           `json:"PLAYERS_COUNT"`
	MyServers       []interface{} `json:"MY_SERVERS"`
	Servers         []struct {
		Name                 string `json:"NAME"`
		IPAddress            string `json:"IP_ADDRESS"`
		Port                 string `json:"PORT"`
		MissionName          string `json:"MISSION_NAME"`
		MissionTime          string `json:"MISSION_TIME"`
		Players              string `json:"PLAYERS"`
		PlayersMax           string `json:"PLAYERS_MAX"`
		Password             string `json:"PASSWORD"`
		Description          string `json:"DESCRIPTION"`
		MissionTimeFormatted string `json:"MISSION_TIME_FORMATTED"`
	} `json:"SERVERS"`
}

// ServerMeta Metadate of servers stored in Redis
type ServerMeta struct {
	Name                 string
	IPAddress            string
	Port                 string
	MissionName          string
	MissionTime          int    // Convert
	Players              int    // Convert
	PlayersMax           int    // Convert
	Password             bool   // "Yes"/"No"
	Description          string // "No"
	MissionTimeFormatted string
	CountryName          string // These are new
	ISO                  string
	LastUpdate           int
}

// GlobalMaxims Maxims recorded for storage in Redis
type GlobalMaxims struct {
	MaxServers     int
	MaxServersTime int
	MaxPlayers     int
	MaxPlayersTime int
}
