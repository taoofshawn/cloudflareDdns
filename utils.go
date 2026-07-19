package main

import (
	"io"
	"log/slog"
	"net/http"
	"time"
)

var ipServiceURL = "https://api.ipify.org"

func getCurrentDynamicIp() string {
	slog.Info("checking current dynamic (origin) IP address")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ipServiceURL)
	if err != nil {
		slog.Error("unable to determine dynamic IP address", "error", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("unable to read IP response", "error", err)
		return ""
	}

	ip := string(body)
	slog.Info("current dynamic IP", "ip", ip)
	return ip
}
