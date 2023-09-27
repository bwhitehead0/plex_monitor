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
*/

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func simpleResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /simpleResponse request\n")
	io.WriteString(w, "This is the simple response.\n")
}

func main() {
	http.HandleFunc("/simpleResponse", simpleResponse)

	err := http.ListenAndServe(":33131", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
