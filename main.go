package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

func main() {

	// Setup log
	flag.Set("logtostderr", "true")
	flag.Parse()
	defer glog.Flush()

	glog.Info("Mornin' Ralph")

	// Collect configuration from environment variables
	config := map[string]string{
		"APITOKEN": os.Getenv("APITOKEN"),
		"NAMELIST": os.Getenv("NAMELIST"),
	}

	for k, v := range config {
		if len(v) == 0 {
			glog.Fatalf("missing environment variable: %s\n", k)
		}
	}

	polltime, err := strconv.Atoi(os.Getenv("POLLTIME"))
	if err != nil {
		polltime = 360
	}

	// Instantiate cloudflare tool
	cfclient := newCloudflareClient(config["APITOKEN"])

	// Format/Prepare list of names frp, collected configuration
	// split comma separated into slice, lowercase, remove whitespace.
	managedNames := strings.Split(
		strings.ToLower(
			strings.Join(
				strings.Fields(config["NAMELIST"]),
				"")),
		",")

	for {
		currentIp := getCurrentDynamicIp()

		for _, managedName := range managedNames {
			record, valuePresent := cfclient.zoneRecords[managedName]
			if !valuePresent {

				glog.Infof("creating new record for: %s", managedName)
				// Create/POST record. define method:
				// cfclient.newRecord(managedName, currentIp)

			} else if currentIp != record.ipAddress {

				glog.Infof("updating record for: %s", managedName)
				// Update/PATCH record. define method:
				// cfclient.updateRecord(managedName, currentIp)

			} else {

				glog.Infof("no update needed for : %s", managedName)
				continue
			}

			//Refresh local data if something changed
			cfclient.getRecords()

		}

		glog.Infof("sleeping for %d minutes", polltime)
		time.Sleep(time.Duration(polltime) * time.Minute)

	}

}
