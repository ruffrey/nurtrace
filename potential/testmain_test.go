package potential

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logLevel := os.Getenv("LOG_LEVEL")
	// disable logs by default in testing
	if logLevel == "" || logLevel == "silent" {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	code := m.Run()
	os.Exit(code)
}
