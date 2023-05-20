package core_test

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/diamondap/vreserve/core"
	"github.com/op/go-logging"
)

func TestInitLogger(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile)
	log, filename := core.InitLogger(logFile, logging.ERROR, false)
	log.Error("Test Message")
	if _, err := os.Stat(logFile); errors.Is(err, os.ErrNotExist) {
		t.Errorf("Log file does not exist at %s", logFile)
	}
	if filename != logFile {
		t.Errorf("Expected log file path '%s', got '%s'", logFile, filename)
	}
	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Error(err)
	}
	if false == strings.HasSuffix(string(data), "Test Message\n") {
		t.Error("Expected message was not in the message log.")
	}
}

func TestDiscardLogger(t *testing.T) {
	log := core.DiscardLogger()
	if log == nil {
		t.Error("DiscardLogger returned nil")
	}
	log.Info("This should not cause an error!")
}
