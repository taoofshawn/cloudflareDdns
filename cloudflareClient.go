package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

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
	validateResponse := Validation{}
	json.Unmarshal(rawToken, &validateResponse)
	for _, message := range validateResponse.Messages {
		glog.Info(message.Message)
	}

	glog.Info("getting zone list")
	rawZones, _ := cfc.call("GET", "zones", "")
	zonesReponse := Zones{}
	json.Unmarshal(rawZones, &zonesReponse)
	for _, result := range zonesReponse.Result {
		glog.Infof("adding %s\n", result.Name)
		cfc.zoneIds = append(cfc.zoneIds, result.ID)
	}

	return &cfc
}

type cloudflareClient struct {
	apiToken    string
	zoneIds     []string
	zoneRecords []zoneRecord
}

func (cfc *cloudflareClient) call(method, resource, data string) ([]byte, error) {
	defer glog.Flush()

	glog.Info("connecting to Cloudflare")

	url := "https://api.cloudflare.com/client/v4/" + resource
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
