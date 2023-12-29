package confluence

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Client for the Confluence API
type Client struct {
	Cookie      string
	Username    string
	Password    string
	AccessToken string
	Endpoint    string
	InsecureTLS bool
	LogLevel    string
}

func (client *Client) request(method string, apiEndpoint string, queryParams string, payload io.Reader, preFns ...PreRequestFn) ([]byte, error) {

	level, err := log.ParseLevel(client.LogLevel)
	if err != nil {
		return nil, err
	}
	log.SetLevel(level)

	if client.InsecureTLS {
		log.Warn("TLS verification is disabled. This allows for man-in-the-middle-attacks.")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	url := client.Endpoint + apiEndpoint

	if queryParams != "" {
		url = url + "?" + queryParams
	}

	log.Trace(fmt.Sprintf("%s %s", method, url))

	req, _ := http.NewRequest(method, url, payload)

	req.Header["X-Atlassian-Token"] = []string{"no-check"}
	req.Header["Content-Type"] = []string{"application/json"}

	for _, preFn := range preFns {
		preFn(req)
	}

	if client.Cookie != "" {
		req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%v", client.Cookie))
	} else if client.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", client.AccessToken))
	} else {
		req.SetBasicAuth(client.Username, client.Password)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("HTTP Request Failed. Received: ", err.Error())
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var errorResponse ErrorResponse
	json.Unmarshal(body, &errorResponse)
	var errMsg string
	for _, s := range errorResponse.Errors {
		if s.Code != "" {
			errMsg = s.Code
		}
		if s.Title != "" {
			errMsg = s.Title
		}
		log.Trace(fmt.Sprintf("[%s] %s %s", s.Code, s.Title, s.Detail))
	}

	switch res.StatusCode {
	case 400:
		return body, fmt.Errorf("error from Confluence API: 400 - bad request: %s", errMsg)
	case 401:
		return body, fmt.Errorf("error from Confluence API: 401 - unauthorized: %s", errMsg)
	case 404:
		return body, fmt.Errorf("error from Confluence API: 404 - not found: %s", errMsg)
	case 413:
		return body, fmt.Errorf("error from Confluence API: 413 - document is too large: %s", errMsg)
	}

	return body, nil
}

// Delete deletes various API types
func (client *Client) Delete(class interface{}) error {
	switch v := class.(type) {
	case Content:
		return client.DeleteContent(class.(Content))
	default:
		return fmt.Errorf("unable to delete type %T", v)
	}
}

// PreRequestFn ...
type PreRequestFn func(request *http.Request)

// QueryParameters provides default query parameters for client
type QueryParameters struct {
	Expand []string `url:"expand,omitempty"`
	Status string   `url:"status,omitempty"`
}

// APIResponse provides default response from API
type APIResponse struct {
	StatusCode int `json:"statusCode,omitempty"`
	Data       struct {
		Authorized bool `json:"authorized,omitempty"`
		Valid      bool `json:"valid,omitempty"`
		Errors     []struct {
			Message struct {
				Key  string        `json:"key,omitempty"`
				Args []interface{} `json:"args,omitempty"`
			} `json:"message,omitempty"`
		} `json:"errors,omitempty"`
		Successful bool `json:"successful,omitempty"`
	} `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Errors []Errors `json:"errors,omitempty"`
}
type Errors struct {
	Status int    `json:"status,omitempty"`
	Code   string `json:"code,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
}
