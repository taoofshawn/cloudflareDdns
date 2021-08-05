package main

import "time"

// https://mholt.github.io/json-to-go/ FTW

// Cloudflare json models

// GET https://api.cloudflare.com/client/v4/user/tokens/verify
type ValidationResponse struct {
	Result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"result"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Type    interface{} `json:"type"`
	} `json:"messages"`
}

// GET https://api.cloudflare.com/client/v4/zones
type ZonesResponse struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []struct {
		ID                  string    `json:"id"`
		Name                string    `json:"name"`
		DevelopmentMode     int       `json:"development_mode"`
		OriginalNameServers []string  `json:"original_name_servers"`
		OriginalRegistrar   string    `json:"original_registrar"`
		OriginalDnshost     string    `json:"original_dnshost"`
		CreatedOn           time.Time `json:"created_on"`
		ModifiedOn          time.Time `json:"modified_on"`
		ActivatedOn         time.Time `json:"activated_on"`
		Owner               struct {
			ID struct {
			} `json:"id"`
			Email struct {
			} `json:"email"`
			Type string `json:"type"`
		} `json:"owner"`
		Account struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"account"`
		Permissions []string `json:"permissions"`
		Plan        struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Price        int    `json:"price"`
			Currency     string `json:"currency"`
			Frequency    string `json:"frequency"`
			LegacyID     string `json:"legacy_id"`
			IsSubscribed bool   `json:"is_subscribed"`
			CanSubscribe bool   `json:"can_subscribe"`
		} `json:"plan"`
		PlanPending struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Price        int    `json:"price"`
			Currency     string `json:"currency"`
			Frequency    string `json:"frequency"`
			LegacyID     string `json:"legacy_id"`
			IsSubscribed bool   `json:"is_subscribed"`
			CanSubscribe bool   `json:"can_subscribe"`
		} `json:"plan_pending"`
		Status      string   `json:"status"`
		Paused      bool     `json:"paused"`
		Type        string   `json:"type"`
		NameServers []string `json:"name_servers"`
	} `json:"result"`
}

// GET https://api.cloudflare.com/client/v4/zones/{ zone ID }/dns_records
type RecordsResponse struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Name       string    `json:"name"`
		Content    string    `json:"content"`
		Proxiable  bool      `json:"proxiable"`
		Proxied    bool      `json:"proxied"`
		TTL        int       `json:"ttl"`
		Locked     bool      `json:"locked"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
		Data       struct {
		} `json:"data"`
		Meta struct {
			AutoAdded bool   `json:"auto_added"`
			Source    string `json:"source"`
		} `json:"meta"`
	} `json:"result"`
}

// PUT/PATCH https://api.cloudflare.com/client/v4/zones/{ zone ID }/dns_records/{ record ID }

type RecordResponse struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Name       string    `json:"name"`
		Content    string    `json:"content"`
		Proxiable  bool      `json:"proxiable"`
		Proxied    bool      `json:"proxied"`
		TTL        int       `json:"ttl"`
		Locked     bool      `json:"locked"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
		Data       struct {
		} `json:"data"`
		Meta struct {
			AutoAdded bool   `json:"auto_added"`
			Source    string `json:"source"`
		} `json:"meta"`
	} `json:"result"`
}
