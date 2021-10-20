package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func compileTemplate(w http.ResponseWriter, title string, content string) {
	type TemplateData struct {
		Title          string
		Content        string
		Active_Servers bool
		Active_Stats   bool
		Active_About   bool
		LastUpdate     string
	}

	var Active_Servers bool
	var Active_Stats bool
	var Active_About bool

	LastUpdate, _ := strconv.ParseInt(redisGet("LastUpdate"), 10, 64)
	t := time.Unix(LastUpdate, 0)

	switch title {
	case "About":
		Active_About = true
	case "Stats":
		Active_Stats = true
	default:
		Active_Servers = true

	}

	data := TemplateData{
		Title:          title,
		Content:        content,
		Active_Servers: Active_Servers,
		Active_Stats:   Active_Stats,
		Active_About:   Active_About,
		LastUpdate:     t.Format(time.RFC3339),
	}

	tmpl := template.Must(template.ParseFiles("templates/template.html"))
	tmpl.Execute(w, data)

}

func handleServers(w http.ResponseWriter, r *http.Request) {

	var err error

	type serversStruct struct {
		ServerCount int
		PlayerCount int
		Servers     []ServerMeta
	}

	// Fetch the data from Redis and Unmarshal it
	metadata := redisGet("metadata")
	//	if metadata == "" {
	//		compileTemplate(w, "Servers", "Server not found.")
	//	}
	var ServerMetaData []ServerMeta
	err = json.Unmarshal([]byte(metadata), &ServerMetaData)
	if err != nil {
		panic(err)
	}

	var playerCount int

	for i, server := range ServerMetaData {
		playerCount += server.Players
		if server.Description == "No" {
			ServerMetaData[i].Description = ""
		} else {
			ServerMetaData[i].Description = strings.ReplaceAll(ServerMetaData[i].Description, "<br />", "")
			ServerMetaData[i].Description = strings.ReplaceAll(ServerMetaData[i].Description, "__", "")
			ServerMetaData[i].Description = strings.ReplaceAll(ServerMetaData[i].Description, "  ", "")
		}
	}

	var sortfields string

	keys, ok := r.URL.Query()["sort"]
	if !ok || len(keys[0]) < 1 {
		sortfields = "Players:DESC"
	} else {
		sortfields = keys[0]
	}

	switch sortfields {
	case "Country:DESC":
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].ISO > ServerMetaData[j].ISO })
	case "Country:ASC":
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].ISO < ServerMetaData[j].ISO })
	case "Server:DESC":
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].Name > ServerMetaData[j].Name })
	case "Server:ASC":
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].Name < ServerMetaData[j].Name })
	case "Scenario:DESC":
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].MissionName > ServerMetaData[j].MissionName })
	case "Scenario:ASC":
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].MissionName < ServerMetaData[j].MissionName })
	case "Players:ASC":
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].Players < ServerMetaData[j].Players })
	default:
		sort.Slice(ServerMetaData, func(i, j int) bool { return ServerMetaData[i].Players > ServerMetaData[j].Players })
	}

	serversContent := serversStruct{len(ServerMetaData), playerCount, ServerMetaData}

	var tpl bytes.Buffer

	parsedTemplate, _ := template.ParseFiles("templates/servers.html")
	err = parsedTemplate.Execute(&tpl, serversContent)
	if err != nil {
		log.Println("Error executing template :", err)
		return
	}

	compileTemplate(w, "Servers", tpl.String())

}

func handleServer(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()

	// Redirect /servers/ -> /servers
	if url == "/servers/" {
		http.Redirect(w, r, "/servers", 301)
	}

	// Parse data
	server := url[9:]
	ServerInfo := redisGet(server)
	var ServerDataMeta ServerMeta
	err := json.Unmarshal([]byte(ServerInfo), &ServerDataMeta)
	//	if err != nil {
	//		log.Println("Server nofound")
	//	}

	var renderRequest string
	renderRequest = "html"
	if r.Header.Get("Accept") == "application/json" {
		renderRequest = "json"
	}

	if renderRequest == "html" {
		var tpl bytes.Buffer
		parsedTemplate, _ := template.ParseFiles("templates/server.html")
		err = parsedTemplate.Execute(&tpl, ServerDataMeta)
		if err != nil {
			log.Println("Error executing template :", err)
			return
		}

		compileTemplate(w, "Servers", tpl.String())
	} else if renderRequest == "json" {
		ServerDataMeta.Description = strings.Replace(ServerDataMeta.Description, "\u003cbr /\u003e", "", -1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ServerDataMeta)
	}

}

func handlerStats(w http.ResponseWriter, r *http.Request) {
	metrics := struct{ stuff int }{stuff: 0}
	var tpl bytes.Buffer
	parsedTemplate, _ := template.ParseFiles("templates/stats.html")
	err := parsedTemplate.Execute(&tpl, metrics)
	if err != nil {
		log.Println("Error executing template :", err)
		return
	}

	compileTemplate(w, "Servers", tpl.String())
}

func handlerAbout(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("templates/about.html")
	if err != nil {
		log.Fatal(err)
	}
	compileTemplate(w, "About", string(content))
}
