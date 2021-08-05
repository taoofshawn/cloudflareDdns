package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

type ZoneRecord struct {
	name       string
	ipAddress  string
	recordId   string
	recordType string
	zoneId     string
}

type cloudflareClient struct {
	apiToken    string
	zoneRecords []ZoneRecord
}

func newCloudflareClient(token string) *cloudflareClient {
	defer glog.Flush()

	cfc := cloudflareClient{
		apiToken: token,
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
	body, _ := ioutil.ReadAll(response.Body)

	return body, err
}

func (cfc *cloudflareClient) getRecords() {

	glog.Info("getting zone list")
	rawZones, _ := cfc.call("GET", "zones", "")
	zonesReponse := ZonesResponse{}
	json.Unmarshal(rawZones, &zonesReponse)

	zoneIds := []string{}
	for _, result := range zonesReponse.Result {
		glog.Infof("adding zone: %s\n", result.Name)
		zoneIds = append(zoneIds, result.ID)
	}

	glog.Info("getting zone records")
	cfc.zoneRecords = nil
	for _, zoneId := range zoneIds {
		rawRecords, _ := cfc.call("GET", "zones/"+zoneId+"/dns_records", "")
		recordsResponse := RecordsResponse{}
		json.Unmarshal(rawRecords, &recordsResponse)

		for _, zoneRecord := range recordsResponse.Result {
			record := ZoneRecord{
				name:       zoneRecord.Name,
				ipAddress:  zoneRecord.Content,
				recordId:   zoneRecord.ID,
				recordType: zoneRecord.Type,
				zoneId:     zoneRecord.ZoneID,
			}
			glog.Infof("adding record: %s\n", record.name)
			cfc.zoneRecords = append(cfc.zoneRecords, record)
		}

	}

}
