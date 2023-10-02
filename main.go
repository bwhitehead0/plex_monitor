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

	TODO:
		- DONE update fmt.Printf, fmt.Println etc to stderr, etc
		- add debug logging
		- add config option to check for updates
		- DONE add config option for port to listen on
		- track duration (receive request, poll API, respond)
		- move startup stuff, sanity checks, etc into init() function
		- note request time in JSON payload response
		- fix http request not updating status (wrapper function?)

*/

package main

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var logger = log.New(os.Stderr, "", 5)
var startTime = time.Now()

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
		logger.Printf("Error reading configuration file %s: %v", file, err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(fileContents, configuration)
	if err != nil {
		logger.Printf("Error parsing configuration file %s: %v", file, err)
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

func response(output string) http.HandlerFunc {
	// using a wrapped handler https://go-cloud-native.com/golang/pass-arguments-to-http-handlers-in-go
	return func(w http.ResponseWriter, r *http.Request) {
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
			logger.Printf("Error connecting to endpoint %s: %s\n", endpoint, err.Error())
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
			logger.Printf("Error connecting to endpoint %s: %s\n", endpoint, err.Error())
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

func main() {
	//time.Sleep(100 * time.Millisecond)
	plexAPIPath := "/identity"
	var endpointStatus string = "Down"
	var decodedResponse plexResponse
	// get config file location from command line argument
	configFile := flag.String("config.file", "", "Config file location")

	flag.Parse()
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

	logger.Printf("Checking API endpoint %s\n", plexAPIEndpoint)

	var requestStart = time.Now()
	plexAPIResponse := pollPlexAPI(plexAPIEndpoint, configuration.IgnoreSSL)
	time.Sleep(4 * time.Second)
	var requestDuration = time.Since(requestStart)

	logger.Printf("API request duration: %s\n", requestDuration)
	if plexAPIResponse != "-1" {
		endpointStatus = "Up"
	}

	xml.Unmarshal([]byte(plexAPIResponse), &decodedResponse)

	plexVersion := strings.Split(decodedResponse.Version, "-")[0]

	// build JSON response
	jsonResponse, err := json.Marshal(map[string]interface{}{
		"Status":          endpointStatus,
		"Version":         plexVersion,
		"RequestDuration": requestDuration,
	})

	logger.Printf("JSON response: %s\n", jsonResponse)

	http.HandleFunc("/status", response(string(jsonResponse)))

	err = http.ListenAndServe(fullListenAddress, nil)

	if errors.Is(err, http.ErrServerClosed) {
		logger.Printf("Server closed\n")
	} else if err != nil {
		logger.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
