package main

// Cloudflare API v4 response models

// --- Common envelope ---

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIMessage struct {
	Message string `json:"message"`
}

// --- Endpoints ---

// GET /user/tokens/verify
type ValidationResponse struct {
	Success  bool         `json:"success"`
	Errors   []APIError   `json:"errors"`
	Messages []APIMessage `json:"messages"`
	Result   *TokenStatus `json:"result"`
}

type TokenStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// GET /zones
type ZonesResponse struct {
	Success    bool         `json:"success"`
	Errors     []APIError   `json:"errors"`
	Messages   []APIMessage `json:"messages"`
	Result     []Zone       `json:"result"`
	ResultInfo *ResultInfo  `json:"result_info"`
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GET /zones/{id}/dns_records
type RecordsResponse struct {
	Success    bool         `json:"success"`
	Errors     []APIError   `json:"errors"`
	Messages   []APIMessage `json:"messages"`
	Result     []DNSRecord  `json:"result"`
	ResultInfo *ResultInfo  `json:"result_info"`
}

type DNSRecord struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	ZoneID   string `json:"zone_id"`
	ZoneName string `json:"zone_name"`
}

// Pagination metadata
type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}
