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
/*

