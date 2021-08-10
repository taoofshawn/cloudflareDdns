package main

// Cloudflare json models (https://mholt.github.io/json-to-go/ FTW)

// GET https://api.cloudflare.com/client/v4/user/tokens/verify

type ValidationResponse struct {
	Messages []struct {
		Message string `json:"message"`
	} `json:"messages"`
}

// GET https://api.cloudflare.com/client/v4/zones
type ZonesResponse struct {
	Result []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}

// GET https://api.cloudflare.com/client/v4/zones/{ zone ID }/dns_records
type RecordsResponse struct {
	Result []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Name     string `json:"name"`
		Content  string `json:"content"`
		ZoneID   string `json:"zone_id"`
		ZoneName string `json:"zone_name"`
	} `json:"result"`
}
