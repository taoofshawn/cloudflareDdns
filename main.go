package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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
		"POLLTIME": os.Getenv("POLLTIME"),
	}

	for k, v := range config {
		if len(v) == 0 {
			glog.Fatalf("missing environment variable: %s\n", k)
		}
	}

	cfclient := newCloudflareClient(config["APITOKEN"])

	// fmt.Println(cfclient.zoneRecords)

	// split comma separated into slice, lowercase, remove whitespace.
	managedNames := strings.Split(
		strings.ToLower(
			strings.Join(
				strings.Fields(config["NAMELIST"]),
				"")),
		",")

	// fmt.Println(names)
	fmt.Println(getCurrentOriginIp())

	for _, managedName := range managedNames {
		fmt.Printf("checking %s\n", managedName)
		matched := false

		for _, record := range cfclient.zoneRecords {

			if managedName == record.name {
				fmt.Printf("will totally update %s\n", record.name)
				matched = true
			}

		}

		if !matched {
			fmt.Printf("Need to create: %s\n", managedName)
		}

	}
}

func getCurrentOriginIp() string {
	defer glog.Flush()
	glog.Info("checking current origin IP address")

	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		glog.Error("unable to determine origin IP address")
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
