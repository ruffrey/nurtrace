package potential

import (
	"io/ioutil"
	"log"
	"os"
)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "silent" {
		log.SetOutput(ioutil.Discard)
	}
}
