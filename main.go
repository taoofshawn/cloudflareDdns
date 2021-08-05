package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
)

func main() {

	flag.Set("logtostderr", "true")
	flag.Parse()
	defer glog.Flush()

	glog.Info("Mornin' Ralph")

	config := map[string]string{
		"APITOKEN": os.Getenv("APITOKEN"),
		"NAMELIST": os.Getenv("NAMELIST"),
	}

	for k, v := range config {
		if len(v) == 0 {
			glog.Fatalf("missing environment variable: %s\n", k)
		}
	}

	cfclient := newCloudflareClient(config["APITOKEN"])

	fmt.Println(cfclient.zoneRecords)
}
