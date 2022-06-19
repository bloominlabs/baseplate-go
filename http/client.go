package http

import (
	"net"
	"net/http"
	"time"
)

// Create new http.Client with sensible defaults including
// 1. SRV lookups with the '.consul' suffix with automatic protocol detection. For example,  https://tip.service.consul will do an SRV lookup
func NewClient() *http.Client {
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: http.DefaultTransport,
	}

	client.Transport.(*http.Transport).DialContext = NewSRVDialer(&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext

	return client
}
