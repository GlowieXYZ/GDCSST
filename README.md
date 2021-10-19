# GDCSST
Digital Combat Simulator Server Tracker written in Go

## About
The goal of GDCSST is to capture server information from the [DCS](digitalcombatsimulator.com/) website, store and visualize it in a human readable format.

## Requirements
To run this use any recent versions of the following:
- [Prometheus](https://github.com/prometheus/prometheus)
- [Redis](https://redis.io/)
- [Prometheus-PNG](https://github.com/lomik/prometheus-png/)

## Installation and Configuration
### 1. Setup a Redis server
This step is easy straightforward, either install one from scratch or reuse an existing instance and give it a new DB for gdcsst.

### 2. Get a copy of the GeoLite2 Free GeoLocation Data
Go to https://dev.maxmind.com/geoip/geolite2-free-geolocation-data?lang=en, register and download `GeoLite2-Country_20yymmdd.tar.gz`. Unpack it and chuck the .mmdb-file in the same folder of the next step.

### 2. Install and run the GDCSST
Copy the binary (you may need to make it with `go build .`) and the `template/` folder to some place nice and comfy where it can run from.

Set the following environmental variables:
#### DCS Account
Login credentials to login at www.digitalcombatsimulator.com and get the list of servers
* `DCS_SERVER_TRACKER_ED_USER`
* `DCS_SERVER_TRACKER_ED_PASS`

#### Redis
Redis server, port and db #
* `DCS_SERVER_TRACKER_REDIS_IP`
* `DCS_SERVER_TRACKER_REDIS_PORT`
* `DCS_SERVER_TRACKER_PASSWORD`
* `DCS_SERVER_TRACKER_REDIS_DB`

#### GeoIP2
Point this to where your GeoLite2-Country.mmdb is stored
* `DCS_SERVER_TRACKER_GEOIP2_FILE`

Finally, chuck these environments in a shell-script of a service file and then run `GDCSST` binary. Once running curl/browse to `host:port/metrics` to see if all goes well, or check the application output.


### 3. Setup Prometheus and point it to your gdcsst install
Install Prometheus and configure a job to scrape GDCSST.
```
scrape_configs:
- job_name: "dcstracker"
  static_configs:
    - targets: ["localhost:9200"]
```

### 4. Install and configure Prometheus-PNG
Download and compile the binary from https://github.com/lomik/prometheus-png/ or run the Docker image:
`docker run --rm -p 8080:8080 lomik/prometheus-png:latest -prometheus "http://127.0.0.1:9090/"`
Pick the right port and point to the right prometheus installation.
