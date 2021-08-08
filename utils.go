package main

import (
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

func getCurrentDynamicIp() string {
	defer glog.Flush()
	glog.Info("checking current Dynamic (origin) IP address")

	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		glog.Error("unable to determine dynamic IP address")
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
