package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	setupSlog()

	slog.Info("Mornin' Ralph")

	config := map[string]string{
		"APITOKEN": os.Getenv("APITOKEN"),
		"NAMELIST": os.Getenv("NAMELIST"),
	}

	for k, v := range config {
		if v == "" {
			slog.Error("missing required environment variable", "var", k)
			os.Exit(1)
		}
	}

	pollMinutes := 60
	if raw := os.Getenv("POLLTIME"); raw != "" {
		if v, err := strconv.Atoi(raw); err != nil {
			slog.Warn("invalid POLLTIME, using default", "value", raw, "default", pollMinutes)
		} else {
			pollMinutes = v
		}
	}

	proxied := true
	if raw := os.Getenv("PROXIED"); raw != "" {
		if v, err := strconv.ParseBool(raw); err != nil {
			slog.Warn("invalid PROXIED, using default", "value", raw, "default", proxied)
		} else {
			proxied = v
		}
	}

	recordType := os.Getenv("RECORD_TYPE")
	if recordType == "" {
		recordType = "A"
	}
	recordType = strings.ToUpper(recordType)
	if recordType != "A" && recordType != "AAAA" {
		slog.Warn("unexpected RECORD_TYPE, falling back to A", "value", recordType)
		recordType = "A"
	}

	cfclient := newCloudflareClient(config["APITOKEN"])

	nameListRaw := strings.Split(config["NAMELIST"], ",")
	managedNames := make([]string, 0, len(nameListRaw))
	for _, name := range nameListRaw {
		name = strings.TrimSpace(strings.ToLower(name))
		if name != "" {
			managedNames = append(managedNames, name)
		}
	}
	if len(managedNames) == 0 {
		slog.Error("NAMELIST contained no valid DNS names")
		os.Exit(1)
	}
	slog.Info("managing records", "names", managedNames, "type", recordType, "proxied", proxied)

	for {
		currentIp := getCurrentDynamicIp()
		if currentIp == "" {
			slog.Error("could not determine current IP, retrying next cycle")
			time.Sleep(time.Duration(pollMinutes) * time.Minute)
			continue
		}

		for _, managedName := range managedNames {
			record, exists := cfclient.zoneRecords[managedName]

			switch {
			case !exists:
				slog.Info("record not found, creating", "name", managedName)
				if err := cfclient.newRecord(managedName, currentIp, recordType, proxied); err != nil {
					slog.Error("failed to create record", "name", managedName, "error", err)
				}

			case currentIp != record.ipAddress:
				slog.Info("IP changed, updating", "name", managedName, "currentIP", currentIp, "oldIP", record.ipAddress)
				if err := cfclient.updateRecord(managedName, currentIp); err != nil {
					slog.Error("failed to update record", "name", managedName, "error", err)
				}

			default:
				slog.Debug("no update needed", "name", managedName)
			}
		}

		slog.Info(fmt.Sprintf("sleeping for %d minutes", pollMinutes))
		time.Sleep(time.Duration(pollMinutes) * time.Minute)
	}
}

func setupSlog() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.String("time", a.Value.Time().Format("2006/01/02 15:04:05"))
			}
			return a
		},
	})
	slog.SetDefault(slog.New(handler))
}
