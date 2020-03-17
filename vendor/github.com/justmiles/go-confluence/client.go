package confluence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Client for the Confluence API
type Client struct {
	Username string
	Password string
	Endpoint string
	Debug    bool
}

func (client *Client) request(method string, apiEndpoint string, queryParams string, payloadString string) ([]byte, error) {
	if client.Debug {
		log.SetLevel(log.DebugLevel)
	}

	var payload io.Reader

	url := client.Endpoint + apiEndpoint

	if queryParams != "" {
		url = url + "?" + queryParams
	}
	if payloadString != "" {
		payload = strings.NewReader(payloadString)
	}

	log.Debug(fmt.Sprintf("%s %s %s", method, url, payloadString))

	req, _ := http.NewRequest(method, url, payload)

	req.Header["X-Atlassian-Token"] = []string{"no-check"}
	req.Header["Content-Type"] = []string{"application/json"}

	req.SetBasicAuth(client.Username, client.Password)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Debugf("Response Status Code: %d", res.StatusCode)
	log.Debugf("Response Body: '%s'", string(body))

	var apiResponse APIResponse

	if string(body) != "" {
		err := json.Unmarshal(body, &apiResponse)
		if err != nil {
			log.Error("Unable to unmarshal API response. Received: '", string(body), "'")
			return body, err
		}

		if apiResponse.Message != "" {
			log.Error(apiResponse.Message)
			if len(apiResponse.Data.Errors) > 0 {
				for _, e := range apiResponse.Data.Errors {
					log.Error("	" + e.Message.Key)
				}
			}
			return body, errors.New(apiResponse.Message)
		}
	}

	return body, nil
}

// PreRequestFn ...
type PreRequestFn func(request *http.Request)

func (client *Client) requestWithFunc(method string, apiEndpoint string, queryParams string, payloadByte *bytes.Buffer, preFn PreRequestFn) ([]byte, error) {
	if client.Debug {
		log.SetLevel(log.DebugLevel)
	}

	var payload io.Reader

	url := client.Endpoint + apiEndpoint

	if queryParams != "" {
		url = url + "?" + queryParams
	}
	if payloadByte != nil {
		payload = payloadByte
	}

	log.Debug(fmt.Sprintf("%s %s", method, url))

	req, _ := http.NewRequest(method, url, payload)

	req.Header["X-Atlassian-Token"] = []string{"no-check"}
	req.Header["Content-Type"] = []string{"application/json"}

	preFn(req)

	req.SetBasicAuth(client.Username, client.Password)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Debugf("Response Status Code: %d", res.StatusCode)
	log.Debugf("Response Body: '%s'", string(body))

	var apiResponse APIResponse

	if string(body) != "" {
		err := json.Unmarshal(body, &apiResponse)
		if err != nil {
			log.Error("Unable to unmarshal API response. Received: '", string(body), "'")
			return body, err
		}

		if apiResponse.Message != "" {
			log.Error(apiResponse.Message)
			if len(apiResponse.Data.Errors) > 0 {
				for _, e := range apiResponse.Data.Errors {
					log.Error("	" + e.Message.Key)
				}
			}
			return body, errors.New(apiResponse.Message)
		}
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
