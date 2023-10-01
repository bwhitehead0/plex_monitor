/*	API endpoint for plex server status

	port 33131

	at minimum should provide the following endpoints:
	-	[url:33131]/status
		reply with HTTP 200 for service up, 500 or service down
	-	DROPPING THIS, /status WILL REPORT THE BELOW

	optional:
	- 	[url:33131]/health
		reply with JSON output for additional info
			- service status (up/down) - based on connection to API endpoint
			- version (example version: 1.32.5.7516-8f4248874 - need to drop from hyphen on)
			- TODO: upgrade available (boolean)
			- TODO: upgrade version available
			- service uptime? DROPPING REQUIREMENT AS NOT NECESSARILY RUNNING LOCALLY

	configuration:
		- address string:		the hostname/IP of the plex server
		- port int:				the port plex runs on
		- ignoressl bool: 		ignore invalid certificate
		- loglevel string:		debug, error, info
		- servicename string: 	plex service name
		- servicecheck bool:	whether to check the service (if not run on plex server)

	TODO:
		- update fmt.Printf, fmt.Println etc to appropriate log level output
		- update service check to accommodate different systems (windows, systemd, init) DROPPING
		- DONE replace TOML config with YAML

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
	"net/http"
	"os"

	// "os/exec"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
	// "honnef.co/go/tools/lintcmd/version"
)

type config struct {
	PlexAddress  string `yaml:"PlexAddress"`
	PlexPort     int    `yaml:"PlexPort"`
	IgnoreSSL    bool   `yaml:"IgnoreSSL"`
	LogLevel     string `yaml:"LogLevel"`
	ServiceName  string `yaml:"ServiceName"`
	ServiceCheck bool   `yaml:"ServiceCheck"`
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
	fileContents, err := os.ReadFile(file)

	if err != nil {
		fmt.Printf("Error reading configuration file %s: %v", file, err)
	}

	err = yaml.Unmarshal(fileContents, configuration)
	if err != nil {
		fmt.Printf("Error parsing configuration file %s: %v", file, err)
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

	fmt.Printf("IgnoreSSL is set to %t\n", ignoreSSL)

	if ignoreSSL {
		// configure to skip TLS verification
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		response, err = client.Get(endpoint)
		if err != nil {
			fmt.Printf("Error connecting to endpoint %s: %s\n", endpoint, err.Error())
			return "-1"
		}
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err.Error())
			return "-1"
		}
		fmt.Println(string(responseData))
		bodyString := string(responseData)
		return bodyString
	} else {
		// validate TLS
		response, err = http.Get(endpoint)
		if err != nil {
			fmt.Printf("Error connecting to endpoint %s: %s\n", endpoint, err.Error())
			return "-1"
		}
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err.Error())
			return "-1"
		}
		fmt.Println(string(responseData))
		bodyString := string(responseData)
		return bodyString
	}
}

func main() {
	plexAPIPath := "/identity"
	var endpointStatus string = "Down"
	var decodedResponse plexResponse
	// get config file location from command line argument
	configFile := flag.String("config.file", "", "Config file location")

	flag.Parse()

	fmt.Println("Using configuration file", *configFile)

	var configuration config
	configuration.readConfig(*configFile)

	fmt.Printf("Configuration data:\n")
	fmt.Printf("%+v\n", configuration)

	// build full API endpoint, convert int port to string with strconv.Itoa
	plexAPIEndpoint := configuration.PlexAddress + ":" + strconv.Itoa(configuration.PlexPort) + plexAPIPath

	fmt.Printf("Checking API endpoint %s\n", plexAPIEndpoint)

	plexAPIResponse := pollPlexAPI(plexAPIEndpoint, configuration.IgnoreSSL)
	if plexAPIResponse != "-1" {
		endpointStatus = "Up"
	}

	xml.Unmarshal([]byte(plexAPIResponse), &decodedResponse)

	plexVersion := strings.Split(decodedResponse.Version, "-")[0]

	// build JSON response
	jsonResponse, err := json.Marshal(map[string]interface{}{
		"Status":  endpointStatus,
		"Version": plexVersion,
	})

	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return
	}

	fmt.Printf("json data: %s\n", jsonResponse)

	http.HandleFunc("/status", response(string(jsonResponse)))

	err = http.ListenAndServe(":33131", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
