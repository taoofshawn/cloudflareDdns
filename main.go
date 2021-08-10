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

	polltime, err := strconv.Atoi(os.Getenv("POLLTIME"))
	if err != nil {
		polltime = 60
	}

	cfclient := newCloudflareClient(config["APITOKEN"])

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
				err := cfclient.newRecord(managedName, currentIp)
				if err != nil {
					glog.Errorf("creating new record for: %s was unsuccessful", managedName)
				}

			} else if currentIp != record.ipAddress {

				glog.Infof("updating record for: %s", managedName)
				err := cfclient.updateRecord(managedName, currentIp)
				if err != nil {
					glog.Errorf("updating record for: %s was unsuccessful ", managedName)
				}
			} else {

				glog.Infof("no update needed for : %s", managedName)
			}

		}

		glog.Infof("sleeping for %d minutes", polltime)
		time.Sleep(time.Duration(polltime) * time.Minute)

	}

}
