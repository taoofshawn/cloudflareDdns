package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

type ZoneRecord struct {
	ipAddress  string
	recordId   string
	recordType string
	zoneId     string
	zoneName   string
}

type cloudflareClient struct {
	apiToken    string
	zoneRecords map[string]ZoneRecord
}

func newCloudflareClient(token string) *cloudflareClient {
	defer glog.Flush()

	cfc := cloudflareClient{
		apiToken:    token,
		zoneRecords: nil,
	}

	glog.Info("verifying token")
	rawToken, err := cfc.call("GET", "user/tokens/verify", "")
	if err != nil {
		glog.Error("unable to verify token.")
	}
	validateResponse := ValidationResponse{}
	json.Unmarshal(rawToken, &validateResponse)
	for _, message := range validateResponse.Messages {
		glog.Info(message.Message)
	}

	cfc.getRecords()

	return &cfc
}

func (cfc *cloudflareClient) call(method, resource, data string) ([]byte, error) {
	defer glog.Flush()

	url := "https://api.cloudflare.com/client/v4/" + resource
	glog.Infof("connecting to Cloudflare: %s", url)
	request, _ := http.NewRequest(method, url, strings.NewReader(data))

	request.Header.Set("Authorization", "Bearer "+cfc.apiToken)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		glog.Error("could not connect to Cloudflare!")
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	return body, err
}

func (cfc *cloudflareClient) getRecords() {
	defer glog.Flush()

	glog.Info("refreshing zone records")

	rawZones, _ := cfc.call("GET", "zones", "")
	zonesReponse := ZonesResponse{}
	json.Unmarshal(rawZones, &zonesReponse)

	zoneIds := []string{}
	for _, result := range zonesReponse.Result {
		glog.Infof("found zone: %s\n", result.Name)
		zoneIds = append(zoneIds, result.ID)
	}

	glog.Info("getting zone records")
	cfc.zoneRecords = make(map[string]ZoneRecord)
	for _, zoneId := range zoneIds {
		rawRecords, _ := cfc.call("GET", "zones/"+zoneId+"/dns_records", "")
		recordsResponse := RecordsResponse{}
		json.Unmarshal(rawRecords, &recordsResponse)

		for _, zoneRecord := range recordsResponse.Result {
			record := ZoneRecord{
				ipAddress:  zoneRecord.Content,
				recordId:   zoneRecord.ID,
				recordType: zoneRecord.Type,
				zoneId:     zoneRecord.ZoneID,
				zoneName:   zoneRecord.ZoneName,
			}
			glog.Infof("found record: %s\n", zoneRecord.Name)
			if record.recordType == "A" {
				cfc.zoneRecords[strings.ToLower(zoneRecord.Name)] = record
			}

		}

	}

}

func (cfc *cloudflareClient) newRecord(managedName string, currentIp string) error {
	defer glog.Flush()

	// Find the zone for the new record
	zone := strings.Join(strings.Split(managedName, ".")[1:], ".")
	zoneId := ""
	zoneExists := false
	for _, record := range cfc.zoneRecords {
		if zone == record.zoneName {
			glog.Infof("record: %s will be added to zone %s", managedName, zone)
			zoneExists = true
			zoneId = record.zoneId
			break
		}
	}

	if !zoneExists {
		glog.Errorf("zone for %s does not exist", managedName)
		return errors.New("bad zone")
	}

	jsonData := fmt.Sprintf(`{
			"type": "A",
			"name": "%s",
			"content": "%s",
			"ttl": 1,
			"proxied": true
		}`, managedName, currentIp)

	response, err := cfc.call("POST", "zones/"+zoneId+"/dns_records", jsonData)
	if err != nil {
		glog.Errorf("\njsonData: %s\nerror: %s\nresponse: %s", jsonData, err, response)
	} else {
		cfc.getRecords()
	}

	return err
}

func (cfc *cloudflareClient) updateRecord(managedName string, currentIp string) error {
	defer glog.Flush()

	zoneId := cfc.zoneRecords[managedName].zoneId
	recordId := cfc.zoneRecords[managedName].recordId
	jsonData := fmt.Sprintf(`{"content": "%s"}`, currentIp)

	response, err := cfc.call("PATCH", "zones/"+zoneId+"/dns_records/"+recordId, jsonData)
	if err != nil {
		glog.Errorf("\njsonData: %s\nerror: %s\nresponse: %s", jsonData, err, response)
	} else {
		cfc.getRecords()
	}

	return err

}
