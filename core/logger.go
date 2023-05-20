package core

import (
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"path/filepath"

	"github.com/op/go-logging"
)

/*
InitLogger creates and returns a logger suitable for logging
human-readable message. filename is the full path to the log
file you want to write. logLevel is the level from go-logging
(DEBUG, INFO, etc.). If logToStderr is true, this will log to
both filename and stderr.
*/
func InitLogger(filename string, logLevel logging.Level, logToStderr bool) (*logging.Logger, string) {
	logDir := filepath.Dir(filename)
	if logDir != "" {
		// If this fails, getRotatingFileWriter will panic in just a second
		_ = os.MkdirAll(logDir, 0755)
	}
	writer, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open log file '%s': %v\n", filename, err)
		os.Exit(1)
	}

	log := logging.MustGetLogger("vreserve")
	format := logging.MustStringFormatter("[%{level}] %{message}")
	logging.SetFormatter(format)
	logging.SetLevel(logLevel, "vreserve")

	logBackend := logging.NewLogBackend(writer, "", stdlog.LstdFlags|stdlog.LUTC)
	if logToStderr {
		// Log to BOTH file and stderr
		stderrBackend := logging.NewLogBackend(os.Stderr, "", stdlog.Lshortfile|stdlog.LstdFlags|stdlog.LUTC)
		stderrBackend.Color = true
		logging.SetBackend(logBackend, stderrBackend)
	} else {
		// Log to file only
		logging.SetBackend(logBackend)
	}

	return log, filename
}

/*
DiscardLogger returns a logger that writes to dev/null.
Suitable for use in testing.
*/
func DiscardLogger() *logging.Logger {
	log := logging.MustGetLogger("vreserve")
	devnull := logging.NewLogBackend(ioutil.Discard, "", 0)
	logging.SetBackend(devnull)
	logging.SetLevel(logging.INFO, "vreserv")
	return log
}

/*
StdoutLogger returns a logger that writes to STDOUT.
*/
func StdoutLogger() *logging.Logger {
	log := logging.MustGetLogger("vreserve")
	logging.SetLevel(logging.INFO, "vreserv")
	return log
}
