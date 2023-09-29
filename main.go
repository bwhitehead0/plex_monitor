/*	API endpoint for plex server status

	port 33131

	at minimum should provide the following endpoints:
	-	[url:33131]/status
		reply with HTTP 200 for service up, 500 or service down

	optional:
	- 	[url:33131]/health
		reply with JSON output for additional info
			- service status (up/down)
			- version
			- upgrade available (boolean)
			- upgrade version available
			- service uptime?

	configuration:
		- address string:		the hostname/IP of the plex server
		- port int:				the port plex runs on
		- ignoressl bool: 		ignore invalid certificate
		- loglevel string:		debug, error, info
		- servicename string: 	plex service name
		- servicecheck bool:	whether to check the service (if not run on plex server)

	TODO:
		- update fmt.Printf, fmt.Println etc to appropriate log level output
		- update service check to accommodate different systems (windows, systemd, init)
		- replace TOML config with YAML

*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
	//"github.com/BurntSushi/toml"
)

type config struct {
	PlexAddress  string `yaml:"PlexAddress"`
	PlexPort     int    `yaml:"PlexPort"`
	IgnoreSSL    bool   `yaml:"IgnoreSSL"`
	LogLevel     string `yaml:"LogLevel"`
	ServiceName  string `yaml:"ServiceName"`
	ServiceCheck bool   `yaml:"ServiceCheck"`
}

func (configuration *config) readConfig(file string) *config {
	// receiver function for configuration file, allows method readConfig(), ie configuration.readConfig()
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

func checkServiceSimple(service string) string {
	// use os.exec to poll 'systemctl check' for service status
	fmt.Printf("Checking status of service %s", service)

	cmd := exec.Command("systemctl", "check", service)

	out, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("systemctl finished with non-zero: %v\n", exitErr)
		} else {
			fmt.Printf("failed to run systemctl: %v", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Service [%v] status: %s\n", service, string(out))
	return string(out)
}

// func simpleResponse(w http.ResponseWriter, r *http.Request, status string) {
// 	fmt.Printf("got /status request\n")
// 	io.WriteString(w, "This is the simple response.\n")
// 	io.WriteString(w, status)
// }

func response(output string) http.HandlerFunc {
	// using a wrapped handler https://go-cloud-native.com/golang/pass-arguments-to-http-handlers-in-go
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, output)
	}
}

func main() {
	// get config file location from command line argument
	configFile := flag.String("config.file", "", "Config file location")

	flag.Parse()

	fmt.Println("Using configuration file", *configFile)

	var configuration config
	configuration.readConfig(*configFile)

	//_, err := toml.Decode(*configFile, &configuration)

	fmt.Printf("Configuration data:\n")
	fmt.Println(configuration)

	var serviceName string = "plexmediaserver"

	serviceStatus := checkServiceSimple(serviceName)
	http.HandleFunc("/status", response(serviceStatus))

	// response(serviceStatus)

	err := http.ListenAndServe(":33131", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
