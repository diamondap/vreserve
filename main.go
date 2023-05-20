package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/diamondap/vreserve/core"
	"github.com/op/go-logging"
)

func main() {
	host, port, logger := parseFlags()
	volumeService := core.NewVolumeService(host, port, logger)
	logger.Infof("vreserv is listening on %s:%d", host, port)
	logger.Infof("To test: curl http://%s:%d/ping", host, port)
	volumeService.Serve()
}

func parseFlags() (string, int, *logging.Logger) {
	var host = flag.String("H", "127.0.0.1", "host to listen on (default 127.0.0.1)")
	var port = flag.Int("p", 8188, "port to listen on (default 8188)")
	var logFile = flag.String("l", "", "path to log file (default STDOUT)")
	var help = flag.Bool("h", false, "print help")
	flag.Parse()
	if *help {
		printUsage()
		os.Exit(0)
	}
	var logger *logging.Logger
	if *logFile == "" {
		logger = core.StdoutLogger()
	} else {
		fmt.Println("vreserve is listening on host", *host, "port", *port)
		fmt.Println("and logging to", *logFile)
		fmt.Println("Use Ctrl-C to stop")
		logger, _ = core.InitLogger(*logFile, logging.INFO, false)
	}
	return *host, *port, logger
}

func printUsage() {
	message := `
vreserve keeps track of how much disk space we have in our local system
Other services reserve space with the vreserve before starting tasks that
require large amounts of disk space. Use Control-C, SIGINT, or SIGKILL to 
shut down the service.

Usage: vreserve [-H=<host>] [-p=<port>] [-l=<log_file]

  - H (host) can be 127.0.0.1 to accept only local requests, 
    or 0.0.0.0 to respond to both local and external requests.
	Default is 127.0.0.1

  - p (port) is the port to listen on. Default is 8188

  - l (log) is the path to the log file. Default is STDOUT

  - h (help) prints this help message

For full documentation, see https://github.com/diamondap/vreserve/README.md

`
	fmt.Println(message)
}
