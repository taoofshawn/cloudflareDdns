package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentDynamicIp_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "203.0.113.42")
	}))
	t.Cleanup(ts.Close)

	orig := ipServiceURL
	ipServiceURL = ts.URL
	t.Cleanup(func() { ipServiceURL = orig })

	ip := getCurrentDynamicIp()
	if ip != "203.0.113.42" {
		t.Errorf("expected 203.0.113.42, got %q", ip)
	}
}

func TestGetCurrentDynamicIp_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(ts.Close)

	orig := ipServiceURL
	ipServiceURL = ts.URL
	t.Cleanup(func() { ipServiceURL = orig })

	ip := getCurrentDynamicIp()
	if ip != "" {
		t.Errorf("expected empty string on server error, got %q", ip)
	}
}

func TestGetCurrentDynamicIp_ConnectionRefused(t *testing.T) {
	orig := ipServiceURL
	ipServiceURL = "http://127.0.0.1:1"
	t.Cleanup(func() { ipServiceURL = orig })

	ip := getCurrentDynamicIp()
	if ip != "" {
		t.Errorf("expected empty string on connection error, got %q", ip)
	}
}
