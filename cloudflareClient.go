package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const cloudflareBaseURL = "https://api.cloudflare.com/client/v4/"

type ZoneRecord struct {
	ipAddress  string
	recordId   string
	recordType string
	zoneId     string
	zoneName   string
}

type cloudflareClient struct {
	apiToken    string
	baseURL     string
	httpClient  *http.Client
	zones       map[string]string // zoneName -> zoneId
	zoneRecords map[string]ZoneRecord
}

func newCloudflareClient(token string) *cloudflareClient {
	cfc := &cloudflareClient{
		apiToken:   token,
		baseURL:    cloudflareBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	slog.Info("verifying token")
	raw, err := cfc.call("GET", "user/tokens/verify", "")
	if err != nil {
		slog.Warn("token verification request failed, continuing", "error", err)
	} else {
		var vr ValidationResponse
		if err := json.Unmarshal(raw, &vr); err != nil {
			slog.Warn("token verification response unparseable", "error", err)
		} else {
			for _, msg := range vr.Messages {
				slog.Info(msg.Message)
			}
			if !vr.Success {
				slog.Warn("token verification returned unsuccessful", "errors", vr.Errors)
			}
		}
	}

	cfc.getZones()
	cfc.getRecords()

	return cfc
}

func (cfc *cloudflareClient) call(method, resource, data string) ([]byte, error) {
	url := cfc.baseURL + resource
	slog.Debug("cf api call", "method", method, "url", url)

	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfc.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return body, nil
}

func (cfc *cloudflareClient) getZones() {
	slog.Info("fetching zones")

	cfc.zones = make(map[string]string)
	page := 1

	for {
		raw, err := cfc.call("GET", fmt.Sprintf("zones?page=%d&per_page=100", page), "")
		if err != nil {
			slog.Error("fetching zones failed", "page", page, "error", err)
			return
		}

		var resp ZonesResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			slog.Error("parsing zones response failed", "error", err)
			return
		}

		for _, z := range resp.Result {
			cfc.zones[z.Name] = z.ID
			slog.Info("found zone", "name", z.Name)
		}

		if resp.ResultInfo == nil || page >= resp.ResultInfo.TotalPages {
			break
		}
		page++
	}
}

func (cfc *cloudflareClient) getRecords() {
	slog.Info("refreshing zone records")

	cfc.zoneRecords = make(map[string]ZoneRecord)

	for zoneName, zoneID := range cfc.zones {
		page := 1

		for {
			raw, err := cfc.call("GET", fmt.Sprintf("zones/%s/dns_records?page=%d&per_page=100", zoneID, page), "")
			if err != nil {
				slog.Error("fetching dns records failed", "zone", zoneName, "error", err)
				break
			}

			var resp RecordsResponse
			if err := json.Unmarshal(raw, &resp); err != nil {
				slog.Error("parsing dns records response failed", "zone", zoneName, "error", err)
				break
			}

			for _, r := range resp.Result {
				if r.Type != "A" {
					continue
				}
				cfc.zoneRecords[strings.ToLower(r.Name)] = ZoneRecord{
					ipAddress:  r.Content,
					recordId:   r.ID,
					recordType: r.Type,
					zoneId:     r.ZoneID,
					zoneName:   r.ZoneName,
				}
				slog.Debug("found a record", "name", r.Name, "ip", r.Content)
			}

			if resp.ResultInfo == nil || page >= resp.ResultInfo.TotalPages {
				break
			}
			page++
		}
	}
}

func (cfc *cloudflareClient) newRecord(managedName, currentIP, recordType string, proxied bool) error {
	zoneName, zoneID := cfc.findZone(managedName)
	if zoneID == "" {
		return fmt.Errorf("zone for %s does not exist", managedName)
	}
	slog.Info("creating new record", "name", managedName, "zone", zoneName, "type", recordType)

	jsonData := fmt.Sprintf(`{
		"type": "%s",
		"name": "%s",
		"content": "%s",
		"ttl": 1,
		"proxied": %t
	}`, recordType, managedName, currentIP, proxied)

	response, err := cfc.call("POST", "zones/"+zoneID+"/dns_records", jsonData)
	if err != nil {
		slog.Error("creating record failed", "name", managedName, "error", err)
		return err
	}

	slog.Info("record created, refreshing", "name", managedName)
	cfc.getRecords()
	_ = response
	return nil
}

func (cfc *cloudflareClient) updateRecord(managedName, currentIP string) error {
	record, ok := cfc.zoneRecords[managedName]
	if !ok {
		return fmt.Errorf("record %s not found in cache", managedName)
	}

	slog.Info("updating record", "name", managedName, "from", record.ipAddress, "to", currentIP)

	jsonData := fmt.Sprintf(`{"content": "%s"}`, currentIP)
	response, err := cfc.call("PATCH", "zones/"+record.zoneId+"/dns_records/"+record.recordId, jsonData)
	if err != nil {
		slog.Error("updating record failed", "name", managedName, "error", err)
		return err
	}

	slog.Info("record updated, refreshing", "name", managedName)
	cfc.getRecords()
	_ = response
	return nil
}

// findZone finds the best-matching zone for a given DNS name.
// For "www.example.com" it matches zone "example.com";
// for "example.com" it matches zone "example.com".
func (cfc *cloudflareClient) findZone(name string) (string, string) {
	name = strings.ToLower(name)
	for zn, zid := range cfc.zones {
		zn = strings.ToLower(zn)
		if name == zn || strings.HasSuffix(name, "."+zn) {
			return zn, zid
		}
	}
	return "", ""
}
