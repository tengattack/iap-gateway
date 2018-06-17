package common

import (
	"io/ioutil"
	"net/http"
)

const (
	// UserAgent is our http request user agent
	UserAgent = "iap-gateway"
)

type MyResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// InitHTTPRequest helps to set necessary headers
func InitHTTPRequest(req *http.Request, isJSONResponse bool) {
	if isJSONResponse {
		req.Header.Set("Accept", "application/json")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
}

// DoHTTPRequest sends request and gets response to struct
func DoHTTPRequest(req *http.Request, isJSONResponse bool, client *http.Client) (*MyResponse, error) {
	InitHTTPRequest(req, isJSONResponse)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// close response
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// LogAccess.Debugf("HTTP %s\n%s", resp.Status, body)

	return &MyResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}, err
}
