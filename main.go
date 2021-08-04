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

	glog.Info("Howdy")

	apiToken := os.Getenv("APITOKEN")
	if len(apiToken) == 0 {
		glog.Fatal("missing environment variable: APITOKEN")
	}

	body := newCloudflareClient(apiToken)
	fmt.Println(body.zoneIds)
	fmt.Println(body.zoneRecords)
}
