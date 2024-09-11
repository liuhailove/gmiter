package client

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		MaxConnsPerHost:       400,
		MaxIdleConnsPerHost:   100,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	var httpClient = http.Client{
		Timeout:   time.Duration(3) * time.Second,
		Transport: transport,
	}

	var req, _ = http.NewRequest(http.MethodGet, "http://10.53.73.120/gmiter", nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp.Proto)
	fmt.Println(string(body))

}
