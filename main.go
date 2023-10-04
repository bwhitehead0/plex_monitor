/*	API endpoint for plex server status

	port 33131

	at minimum should provide the following endpoint(s):
	-	[url:33131]/status
		reply with JSON payload
			- service status (up/down) - based on connection to API endpoint
			- version (example version: 1.32.5.7516-8f4248874 - need to drop from hyphen on)
			- TODO: upgrade available (boolean)
			- TODO: upgrade version available

	configuration:
		- address string:		the hostname/IP of the plex server
		- port int:				the port plex runs on
		- ignoressl bool: 		ignore invalid certificate, etc
		- listen port:			TCP port for this API endpoint
		- listen address:		IP address for this API to bind to

	TODO:
		- DONE update fmt.Printf, fmt.Println etc to stderr, etc
		- add debug logging
		- add config option to check for updates
		- DONE add listen address
		- DONE add config option for port to listen on
		- DONE track duration (startup, poll API, etc)
		- move startup stuff, sanity checks, etc into init() function
		- DONE note request time in JSON payload response
		- DONE fix http request not updating status (wrapper function?)
		- DONE move tasks to functions (xml parsing, etc)
		- DONE Intercept ctrl-c/sigint for graceful shutdown.
		- fix startup time formatting? (sometimes in nanoseconds/microseconds)
*/

package main

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

var logger = log.New(os.Stderr, "", 5)
var startTime = time.Now()
var appVersion = "0.1.0"

type config struct {
	PlexAddress   string `yaml:"PlexAddress"`
	PlexPort      int    `yaml:"PlexPort"`
	IgnoreSSL     bool   `yaml:"IgnoreSSL"`
	ListenAddress string `yaml:"ListenAddress"`
	ListenPort    int    `yaml:"ListenPort"`
}

type plexResponse struct {
	// <MediaContainer size="0" claimed="1" machineIdentifier="ee2e37973bc957d96a81bad551adef994763b651" version="1.32.5.7516-8f4248874"> </MediaContainer>
	MediaContainer    string `xml:",chardata"`
	Size              int    `xml:"size,attr"`
	Claimed           int    `xml:"claimed,attr"`
	MachineIdentifier string `xml:"machineIdentifier,attr"`
	Version           string `xml:"version,attr"`
}

func (configuration *config) readConfig(file string) *config {
	// receiver function for configuration file, allows method readConfig(), ie configuration.readConfig(file)
	var defaultAddress string = "0.0.0.0"
	var defaultPort int = 33131
	fileContents, err := os.ReadFile(file)

	if err != nil {
		logger.Printf("Error reading configuration file %s: %v\n", file, err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(fileContents, configuration)
	if err != nil {
		logger.Printf("Error parsing configuration file %s: %v\n", file, err)
		os.Exit(1)
	}

	if configuration.ListenAddress == "" {
		configuration.ListenAddress = defaultAddress
		logger.Printf("Using default listen Address %v\n", configuration.ListenAddress)
	}

	if configuration.ListenPort == 0 {
		configuration.ListenPort = defaultPort
		logger.Printf("Using default listen port %v\n", configuration.ListenPort)
	}

	return configuration
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func response(endpoint string, ignoreSSL bool) http.HandlerFunc {
	// using a wrapped handler https://go-cloud-native.com/golang/pass-arguments-to-http-handlers-in-go
	return func(w http.ResponseWriter, r *http.Request) {
		sourceIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			logger.Printf("Error getting client IP.\n")
		}
		output := getResponse(endpoint, ignoreSSL, sourceIP)

		io.WriteString(w, output)
	}
}

func pollPlexAPI(endpoint string, ignoreSSL bool) string {
	// function might need to be reevaluated for efficiency
	var response *http.Response
	var err error

	logger.Printf("IgnoreSSL is set to %t\n", ignoreSSL)

	if ignoreSSL {
		// configure to skip TLS verification
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		response, err = client.Get(endpoint)
		if err != nil {
			logger.Printf("Error connecting to endpoint: %s\n", err.Error())
			return "-1"
		}
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			logger.Printf("Error reading response: %v\n", err.Error())
			return "-1"
		}
		// fmt.Println(string(responseData))
		bodyString := string(responseData)
		return bodyString
	} else {
		// validate TLS
		response, err = http.Get(endpoint)
		if err != nil {
			logger.Printf("Error connecting to endpoint: %s\n", err.Error())
			return "-1"
		}
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			logger.Printf("Error reading response: %v\n", err.Error())
			return "-1"
		}
		// fmt.Println(string(responseData))
		bodyString := string(responseData)
		return bodyString
	}
}

func convertToJson(apiResponse string, requestDuration time.Duration, requestStart time.Time) string {
	var decodedResponse plexResponse
	var endpointStatus string = "Down"

	// apiResponse from pollPlexAPI()
	if apiResponse != "-1" {
		endpointStatus = "Up"
	}

	xml.Unmarshal([]byte(apiResponse), &decodedResponse)

	plexVersion := strings.Split(decodedResponse.Version, "-")[0]

	// build JSON response
	jsonResponse, err := json.Marshal(map[string]interface{}{
		"Status":          endpointStatus,
		"Version":         plexVersion,
		"RequestDuration": requestDuration.Milliseconds(),
		"RequestTime":     requestStart.UTC(),
	})

	if err != nil {
		logger.Printf("Error building response: %s\n", err)
		jsonResponse := "{" + "\"Error\": " + "\"" + err.Error() + "\"}"
		logger.Printf("JSON response: %s\n", jsonResponse)
		return string(jsonResponse)
	}

	logger.Printf("JSON response: %s\n", jsonResponse)
	return string(jsonResponse) + "\n"
}

func getResponse(endpoint string, ignoreSSL bool, sourceIP string) string {
	// continue here
	logger.Printf("Received request for endpoint '/status' from %s\n", sourceIP)
	logger.Printf("Checking API endpoint %s\n", endpoint)
	var requestStart = time.Now()
	apiResponse := pollPlexAPI(endpoint, ignoreSSL)
	var requestDuration = time.Since(requestStart)

	jsonResult := convertToJson(apiResponse, requestDuration, requestStart.UTC())

	return jsonResult
}

func main() {
	plexAPIPath := "/identity"
	// capture sigint, sigterm
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-signals
		//fmt.Println()
		if sig == os.Interrupt {
			// if sigint / ctrl-c
			fmt.Print("\r") // so we don't write ^C to terminal on sigint
		}
		logger.Printf("Received %s signal. Exiting.\n", sig)
		os.Exit(0)
		done <- true
	}()

	// get config file location from command line argument
	configFile := flag.String("config.file", "", "Config file location")

	flag.Parse()

	if !isFlagPassed("config.file") {
		// missing config.file flag
		logger.Printf("Error: configuration file not specified.\n")
		os.Exit(1)
	}

	logger.Printf("Plex Monitor v%s starting up.", appVersion)

	configFullPath, err := filepath.Abs(*configFile)

	if err != nil {
		log.Fatalf("Error finding config file %s: %s\n", *configFile, err)
	}

	logger.Println("Using configuration file", configFullPath)

	var configuration config
	configuration.readConfig(*configFile)

	logger.Printf("Startup time elapsed: %s\n", time.Since(startTime))

	var fullListenAddress = configuration.ListenAddress + ":" + strconv.Itoa(configuration.ListenPort)

	// build full API endpoint, convert int port to string with strconv.Itoa
	plexAPIEndpoint := configuration.PlexAddress + ":" + strconv.Itoa(configuration.PlexPort) + plexAPIPath

	http.HandleFunc("/status", response(plexAPIEndpoint, configuration.IgnoreSSL))

	err = http.ListenAndServe(fullListenAddress, nil)

	if errors.Is(err, http.ErrServerClosed) {
		logger.Printf("Server closed\n")
	} else if err != nil {
		logger.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}

	<-done
}
