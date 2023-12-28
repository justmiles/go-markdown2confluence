package confluence

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	log "github.com/sirupsen/logrus"
)

// GetSpaces returns all spaces
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-space/#api-spaces-get
func (client *Client) GetSpaces(qp *GetSpacesQueryParameters) (GetSpacesResponse, error) {

	v, _ := query.Values(qp)
	queryParams := v.Encode()

	body, err := client.request("GET", "/api/v2/spaces", queryParams, nil)
	if err != nil {
		return GetSpacesResponse{}, err
	}

	var response GetSpacesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error("Unable to unmarshal response. Received: '", string(body), "'")
	}

	return response, err
}

func (client *Client) GetSpaceByKey(key string) (*Space, error) {

	res, err := client.GetSpaces(&GetSpacesQueryParameters{Keys: []string{key}})
	if err != nil {
		return nil, err
	}

	if len(res.Results) == 0 {
		return nil, fmt.Errorf("space not found")
	}

	return &res.Results[0], err
}

// UpdateSpaceHomePage updates the homepage of a given space
// https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-space/#api-wiki-rest-api-space-spacekey-put
func (client *Client) UpdateSpaceHomePage(key string, pageID string) error {
	payload := fmt.Sprintf(`{ "homepage": { "id": "%s" } }`, pageID)

	_, err := client.request("PUT", fmt.Sprintf("/rest/api/space/%s", key), "", strings.NewReader(payload))
	if err != nil {
		return err
	}

	return err
}

type GetSpacesQueryParameters struct {
	Keys   []string `url:"keys,omitempty"`   // Filter the results to spaces based on their keys. Multiple keys can be specified as a comma-separated list.
	Status string   `url:"status,omitempty"` // Filter the results to spaces based on their keys. Multiple keys can be specified as a comma-separated list.
}

type GetSpacesResponse struct {
	Results []Space `json:"results,omitempty"`
	Links   Links   `json:"_links,omitempty"`
}

type Links struct {
	Webui    string `json:"webui,omitempty"`
	Next     string `json:"next,omitempty"`
	Editui   string `json:"editui,omitempty"`
	Tinyui   string `json:"tinyui,omitempty"`
	Download string `json:"download,omitempty"`
}

type Space struct {
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	AuthorID    string    `json:"authorId,omitempty"`
	HomepageID  string    `json:"homepageId,omitempty"`
	Icon        any       `json:"icon,omitempty"`
	Name        string    `json:"name,omitempty"`
	Key         string    `json:"key,omitempty"`
	ID          string    `json:"id,omitempty"`
	Type        string    `json:"type,omitempty"`
	Description any       `json:"description,omitempty"`
	Status      string    `json:"status,omitempty"`
	Links       Links     `json:"_links,omitempty"`
}

// json returns the json representation of a space
func (s *Space) json() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
