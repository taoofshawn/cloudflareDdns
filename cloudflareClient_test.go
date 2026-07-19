package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockCloudflare returns an httptest.Server that responds to Cloudflare API paths.
func mockCloudflare(
	t *testing.T,
	verifyHandler func(w http.ResponseWriter, r *http.Request),
	zonesHandler func(w http.ResponseWriter, r *http.Request),
	recordsHandler func(w http.ResponseWriter, r *http.Request),
) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/client/v4/user/tokens/verify", func(w http.ResponseWriter, r *http.Request) {
		if verifyHandler != nil {
			verifyHandler(w, r)
		} else {
			json.NewEncoder(w).Encode(ValidationResponse{
				Success:  true,
				Messages: []APIMessage{{Message: "token valid"}},
				Result:   &TokenStatus{ID: "1", Status: "active"},
			})
		}
	})

	mux.HandleFunc("/client/v4/zones", func(w http.ResponseWriter, r *http.Request) {
		if zonesHandler != nil {
			zonesHandler(w, r)
		} else {
			json.NewEncoder(w).Encode(ZonesResponse{
				Success: true,
				Result: []Zone{
					{ID: "zone1", Name: "example.com"},
				},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
			})
		}
	})

	mux.HandleFunc("/client/v4/zones/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/dns_records") {
			if recordsHandler != nil {
				recordsHandler(w, r)
			} else {
				json.NewEncoder(w).Encode(RecordsResponse{
					Success: true,
					Result: []DNSRecord{
						{ID: "rec1", Type: "A", Name: "www.example.com", Content: "1.2.3.4", ZoneID: "zone1", ZoneName: "example.com"},
					},
					ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
				})
			}
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("unhandled request: %s %s", r.Method, r.URL.Path)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "result": nil})
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func newTestClient(t *testing.T, baseURL string) *cloudflareClient {
	t.Helper()
	cfc := &cloudflareClient{
		apiToken:   "test-token",
		baseURL:    baseURL + "/client/v4/",
		httpClient: &http.Client{},
	}
	cfc.getZones()
	cfc.getRecords()
	return cfc
}

func TestNewCloudflareClient_TokenVerifySuccess(t *testing.T) {
	ts := mockCloudflare(t, nil, nil, nil)
	_ = newTestClient(t, ts.URL)
}

func TestNewCloudflareClient_TokenVerifyFails(t *testing.T) {
	ts := mockCloudflare(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		nil, nil,
	)

	cfc := newTestClient(t, ts.URL)
	if _, ok := cfc.zones["example.com"]; !ok {
		t.Error("expected zones to be populated despite token failure")
	}
	if rec, ok := cfc.zoneRecords["www.example.com"]; !ok {
		t.Error("expected records to be populated despite token failure")
	} else if rec.ipAddress != "1.2.3.4" {
		t.Errorf("expected ip 1.2.3.4, got %s", rec.ipAddress)
	}
}

func TestGetRecords_Pagination(t *testing.T) {
	pageCalls := 0
	ts := mockCloudflare(t, nil, nil,
		func(w http.ResponseWriter, r *http.Request) {
			pageCalls++
			if pageCalls == 1 {
				json.NewEncoder(w).Encode(RecordsResponse{
					Success: true,
					Result: []DNSRecord{
						{ID: "r1", Type: "A", Name: "page1.example.com", Content: "1.1.1.1", ZoneID: "zone1", ZoneName: "example.com"},
					},
					ResultInfo: &ResultInfo{Page: 1, TotalPages: 2},
				})
			} else {
				json.NewEncoder(w).Encode(RecordsResponse{
					Success: true,
					Result: []DNSRecord{
						{ID: "r2", Type: "A", Name: "page2.example.com", Content: "2.2.2.2", ZoneID: "zone1", ZoneName: "example.com"},
					},
					ResultInfo: &ResultInfo{Page: 2, TotalPages: 2},
				})
			}
		},
	)

	cfc := newTestClient(t, ts.URL)
	if pageCalls != 2 {
		t.Errorf("expected 2 page calls, got %d", pageCalls)
	}
	if _, ok := cfc.zoneRecords["page2.example.com"]; !ok {
		t.Error("expected record from page 2 to be present")
	}
}

func TestGetRecords_FiltersOnlyARecords(t *testing.T) {
	ts := mockCloudflare(t, nil, nil,
		func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(RecordsResponse{
				Success: true,
				Result: []DNSRecord{
					{ID: "r1", Type: "A", Name: "www.example.com", Content: "1.1.1.1", ZoneID: "zone1", ZoneName: "example.com"},
					{ID: "r2", Type: "AAAA", Name: "www.example.com", Content: "::1", ZoneID: "zone1", ZoneName: "example.com"},
					{ID: "r3", Type: "CNAME", Name: "blog.example.com", Content: "other.example.com", ZoneID: "zone1", ZoneName: "example.com"},
					{ID: "r4", Type: "MX", Name: "example.com", Content: "mail.example.com", ZoneID: "zone1", ZoneName: "example.com"},
				},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
			})
		},
	)

	cfc := newTestClient(t, ts.URL)
	if len(cfc.zoneRecords) != 1 {
		t.Errorf("expected exactly 1 A record, got %d", len(cfc.zoneRecords))
	}
	if _, ok := cfc.zoneRecords["www.example.com"]; !ok {
		t.Error("expected www.example.com A record to be present")
	}
}

func TestGetRecords_MultipleZones(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/client/v4/user/tokens/verify", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ValidationResponse{Success: true, Messages: []APIMessage{{Message: "ok"}}})
	})

	mux.HandleFunc("/client/v4/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ZonesResponse{
			Success: true,
			Result: []Zone{
				{ID: "zone1", Name: "example.com"},
				{ID: "zone2", Name: "example.org"},
			},
			ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
		})
	})

	mux.HandleFunc("/client/v4/zones/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/dns_records") {
			return
		}
		if strings.Contains(r.URL.Path, "zone1") {
			json.NewEncoder(w).Encode(RecordsResponse{
				Success: true,
				Result: []DNSRecord{
					{ID: "r1", Type: "A", Name: "www.example.com", Content: "1.1.1.1", ZoneID: "zone1", ZoneName: "example.com"},
				},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
			})
		} else {
			json.NewEncoder(w).Encode(RecordsResponse{
				Success: true,
				Result: []DNSRecord{
					{ID: "r2", Type: "A", Name: "www.example.org", Content: "2.2.2.2", ZoneID: "zone2", ZoneName: "example.org"},
				},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
			})
		}
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	cfc := &cloudflareClient{
		apiToken:   "test-token",
		baseURL:    ts.URL + "/client/v4/",
		httpClient: &http.Client{},
	}
	cfc.getZones()
	cfc.getRecords()

	if len(cfc.zoneRecords) != 2 {
		t.Errorf("expected 2 total records, got %d", len(cfc.zoneRecords))
	}
}

func TestNewRecord_CreatesNewRecord(t *testing.T) {
	var capturedBody string
	ts := mockCloudflare(t, nil, nil,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				buf := make([]byte, r.ContentLength)
				r.Body.Read(buf)
				capturedBody = string(buf)

				json.NewEncoder(w).Encode(RecordsResponse{
					Success:    true,
					Result:     []DNSRecord{},
					ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
				})
				return
			}
			json.NewEncoder(w).Encode(RecordsResponse{
				Success: true,
				Result: []DNSRecord{
					{ID: "rec1", Type: "A", Name: "www.example.com", Content: "1.2.3.4", ZoneID: "zone1", ZoneName: "example.com"},
				},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
			})
		},
	)

	cfc := newTestClient(t, ts.URL)

	err := cfc.newRecord("sub.example.com", "5.6.7.8", "A", false)
	if err != nil {
		t.Fatalf("newRecord failed: %v", err)
	}

	if !strings.Contains(capturedBody, `"type": "A"`) {
		t.Errorf("expected type A: %s", capturedBody)
	}
	if !strings.Contains(capturedBody, `"content": "5.6.7.8"`) {
		t.Errorf("expected content 5.6.7.8: %s", capturedBody)
	}
	if !strings.Contains(capturedBody, `"proxied": false`) {
		t.Errorf("expected proxied false: %s", capturedBody)
	}
}

func TestNewRecord_ApexDomain(t *testing.T) {
	ts := mockCloudflare(t, nil, nil,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				json.NewEncoder(w).Encode(RecordsResponse{Success: true, Result: []DNSRecord{}, ResultInfo: &ResultInfo{Page: 1, TotalPages: 1}})
				return
			}
			json.NewEncoder(w).Encode(RecordsResponse{
				Success:    true,
				Result:     []DNSRecord{{ID: "apex", Type: "A", Name: "example.com", Content: "1.1.1.1", ZoneID: "zone1", ZoneName: "example.com"}},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
			})
		},
	)

	cfc := newTestClient(t, ts.URL)

	err := cfc.newRecord("example.com", "9.9.9.9", "A", true)
	if err != nil {
		t.Errorf("newRecord for apex domain should succeed, got: %v", err)
	}
}

func TestNewRecord_ZoneNotFound(t *testing.T) {
	ts := mockCloudflare(t, nil, nil, nil)
	cfc := newTestClient(t, ts.URL)

	err := cfc.newRecord("nonexistent.remote", "1.2.3.4", "A", true)
	if err == nil {
		t.Fatal("expected error for unknown zone")
	}
}

func TestUpdateRecord_SendsPatch(t *testing.T) {
	method := ""
	var capturedBody string
	ts := mockCloudflare(t, nil, nil,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPatch {
				method = "PATCH"
				buf := make([]byte, r.ContentLength)
				r.Body.Read(buf)
				capturedBody = string(buf)
			}
			json.NewEncoder(w).Encode(RecordsResponse{
				Success:    true,
				Result:     []DNSRecord{{ID: "rec1", Type: "A", Name: "www.example.com", Content: "5.6.7.8", ZoneID: "zone1", ZoneName: "example.com"}},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 1},
			})
		},
	)

	cfc := newTestClient(t, ts.URL)

	err := cfc.updateRecord("www.example.com", "5.6.7.8")
	if err != nil {
		t.Fatalf("updateRecord failed: %v", err)
	}

	if method != "PATCH" {
		t.Errorf("expected PATCH request, got %s", method)
	}
	if !strings.Contains(capturedBody, `"content": "5.6.7.8"`) {
		t.Errorf("expected content 5.6.7.8 in body: %s", capturedBody)
	}
}

func TestUpdateRecord_NotFound(t *testing.T) {
	ts := mockCloudflare(t, nil, nil, nil)
	cfc := newTestClient(t, ts.URL)

	err := cfc.updateRecord("missing.example.com", "1.2.3.4")
	if err == nil {
		t.Fatal("expected error for unknown record")
	}
}

func TestFindZone(t *testing.T) {
	cfc := &cloudflareClient{
		zones: map[string]string{
			"example.com":     "zone1",
			"sub.example.net": "zone2",
		},
	}

	tests := []struct {
		name     string
		expected string
		wantOk   bool
	}{
		{"www.example.com", "zone1", true},
		{"example.com", "zone1", true},
		{"deep.sub.example.com", "zone1", true},
		{"other.org", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zn, zid := cfc.findZone(tt.name)
			if tt.wantOk {
				if zid != tt.expected {
					t.Errorf("findZone(%q) = (%q, %q), want zoneID %q", tt.name, zn, zid, tt.expected)
				}
			} else if zid != "" {
				t.Errorf("findZone(%q) = (%q, %q), wanted no match", tt.name, zn, zid)
			}
		})
	}
}

func TestCall_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {} // hang forever
	}))
	t.Cleanup(ts.Close)

	cfc := &cloudflareClient{
		apiToken:   "test",
		baseURL:    ts.URL + "/",
		httpClient: &http.Client{Timeout: 1},
	}

	_, err := cfc.call("GET", "test", "")
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestCall_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, `{"success":false,"errors":[{"code":9103,"message":"bad token"}]}`)
	}))
	t.Cleanup(ts.Close)

	cfc := &cloudflareClient{
		apiToken:   "bad",
		baseURL:    ts.URL + "/",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	body, err := cfc.call("GET", "test", "")
	if err != nil {
		t.Fatalf("HTTP errors should still return the body: %v", err)
	}

	var resp struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false for bad token")
	}
}

func TestGetZones_Pagination(t *testing.T) {
	pageCalls := 0
	mux := http.NewServeMux()

	mux.HandleFunc("/client/v4/user/tokens/verify", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ValidationResponse{Success: true})
	})

	mux.HandleFunc("/client/v4/zones", func(w http.ResponseWriter, r *http.Request) {
		pageCalls++
		if pageCalls == 1 {
			json.NewEncoder(w).Encode(ZonesResponse{
				Success:    true,
				Result:     []Zone{{ID: "z1", Name: fmt.Sprintf("zone%d.example.com", pageCalls)}},
				ResultInfo: &ResultInfo{Page: 1, TotalPages: 3},
			})
		} else if pageCalls == 2 {
			json.NewEncoder(w).Encode(ZonesResponse{
				Success:    true,
				Result:     []Zone{{ID: "z2", Name: fmt.Sprintf("zone%d.example.com", pageCalls)}},
				ResultInfo: &ResultInfo{Page: 2, TotalPages: 3},
			})
		} else {
			json.NewEncoder(w).Encode(ZonesResponse{
				Success:    true,
				Result:     []Zone{{ID: "z3", Name: fmt.Sprintf("zone%d.example.com", pageCalls)}},
				ResultInfo: &ResultInfo{Page: 3, TotalPages: 3},
			})
		}
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	cfc := &cloudflareClient{
		apiToken:   "test",
		baseURL:    ts.URL + "/client/v4/",
		httpClient: &http.Client{},
	}
	cfc.getZones()

	if pageCalls != 3 {
		t.Errorf("expected 3 zone page calls, got %d", pageCalls)
	}
	if len(cfc.zones) != 3 {
		t.Errorf("expected 3 zones, got %d", len(cfc.zones))
	}
}
