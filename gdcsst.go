package main

// Imports
import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var AppConfig Config

var (
	// GdcsstOnline Online servers or players
	GdcsstOnline = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gdcsst_online",
		Help: "Online",
	},
		[]string{"scope"},
	)

	// GdcsstPlayers -- Players on Servers
	GdcsstPlayers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gdcsst_players",
		Help: "All players on a server",
	},
		[]string{"server"},
	)
)

func init() {
	prometheus.MustRegister(GdcsstOnline)
	prometheus.MustRegister(GdcsstPlayers)
	AppConfig = processConfig()
}

// Required for Redis
var ctx = context.Background()

// Main function
func main() {
	// The Main loop to update servers
	// Handled in "backend.go"
	go updateServers()

	// Configure the net/http module
	// Handled in "frontend.go"
	http.Handle("/", http.RedirectHandler("/servers", 301))
	http.HandleFunc("/servers", handleServers)
	http.HandleFunc("/servers/", handleServer)
	http.HandleFunc("/stats", handlerStats)
	http.HandleFunc("/about", handlerAbout)

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	log.Fatal(http.ListenAndServe(AppConfig.ListenAddress, nil))
}
