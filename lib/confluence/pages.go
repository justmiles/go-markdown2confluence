package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-querystring/query"
	log "github.com/sirupsen/logrus"
)

type GetPagesInSpaceQueryParameters struct {
	Status     []string `url:"status,omitempty"`      // Filter the results to pages based on their status. By default, current and archived are used.
	BodyFormat string   `url:"body-format,omitempty"` // Maximum number of pages per result to return. If more results exist, use the Link header to retrieve a relative URL that will return the next set of results.
	Title      string   `url:"title,omitempty"`
	Depth      string   `url:"depth,omitempty"`
	GetResultsQueryParameters
}

type GetPagesInSpaceResponse struct {
	Results []Page `json:"results,omitempty"`
	Links   Links  `json:"_links,omitempty"`
}

type PageVersion struct {
	Number    int       `json:"number,omitempty"`
	Message   string    `json:"message,omitempty"`
	MinorEdit bool      `json:"minorEdit,omitempty"`
	AuthorID  string    `json:"authorId,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type Page struct {
	ID             string  `json:"id,omitempty"`
	Status         string  `json:"status,omitempty"`
	Title          string  `json:"title,omitempty"`
	SpaceID        string  `json:"spaceId,omitempty"`
	RemoteParentID string  `json:"parentId,omitempty"`
	ParentType     string  `json:"parentType,omitempty"`
	Position       int     `json:"position,omitempty"`
	AuthorID       string  `json:"authorId,omitempty"`
	OwnerID        string  `json:"ownerId,omitempty"`
	LastOwnerID    string  `json:"lastOwnerId,omitempty"`
	CreatedAt      string  `json:"createdAt,omitempty"`
	Version        Version `json:"version,omitempty"`
	Body           Body    `json:"body,omitempty"`
	Links          Links   `json:"_links,omitempty"`
}

type CreatePageQueryParameters struct {
	Embedded  bool `url:"embedded,omitempty"`   // Tag the content as embedded and content will be created in NCS.
	Private   bool `url:"private,omitempty"`    // The page will be private. Only the user who creates this page will have permission to view and edit one.
	RootLevel bool `url:"root-level,omitempty"` // The page will be created at the root level of the space (outside the space homepage tree).
}

type CreatePageBody struct {
	ID             string      `json:"id,omitempty"`
	SpaceID        string      `json:"spaceId,omitempty"`
	Status         string      `json:"status,omitempty"`
	Title          string      `json:"title,omitempty"`
	RemoteParentID string      `json:"parentId,omitempty"`
	Body           ContentBody `json:"body,omitempty"`
	Version        Version     `json:"version,omitempty"`
}

type ContentBody struct {
	Representation string `json:"representation,omitempty"`
	Value          string `json:"value,omitempty"`
}

type Version struct {
	CreatedAt string `json:"createdAt,omitempty"`
	Message   string `json:"message,omitempty"`
	Number    int    `json:"number,omitempty"`
	MinorEdit bool   `json:"minorEdit,omitempty"`
	AuthorID  string `json:"authorId,omitempty"`
}

type Body struct {
	Storage        ContentBody `json:"storage,omitempty"`
	AtlasDocFormat ContentBody `json:"atlas_doc_format,omitempty"`
	View           ContentBody `json:"view,omitempty"`
}

type GetPagesResponse struct {
	Results []Page `json:"results"`
	Links   Links  `json:"_links"`
}

// GetPagesInSpace Returns all pages
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-spaces-id-pages-get
func (client *Client) GetPagesInSpace(id string, qp *GetPagesInSpaceQueryParameters) (GetPagesInSpaceResponse, error) {
	var response GetPagesInSpaceResponse

	v, _ := query.Values(qp)
	queryParams := v.Encode()

	body, err := client.request("GET", fmt.Sprintf("/api/v2/spaces/%s/pages", id), queryParams, nil)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("unable to marshal response: %s", err)
	}

	return response, err
}

// CreatePage creates a page in the space
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-post
func (client *Client) CreatePage(createPageBody *CreatePageBody, qp *CreatePageQueryParameters) (Page, error) {
	var response Page

	v, _ := query.Values(qp)
	queryParams := v.Encode()

	payload, err := json.Marshal(createPageBody)
	if err != nil {
		log.Error("Unable to marshal body. Received: '", err, "'")
	}

	body, err := client.request("POST", "/api/v2/pages", queryParams, bytes.NewReader(payload))
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("unable to marshal response: %s", err)
	}

	return response, err
}

// UpdatePage update a page by id
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-id-put
func (client *Client) UpdatePage(pageID string, createPageBody *CreatePageBody) (Page, error) {
	var response Page

	payload, err := json.Marshal(createPageBody)
	if err != nil {
		log.Error("Unable to marshal body. Received: '", err, "'")
	}

	body, err := client.request("PUT", fmt.Sprintf("/api/v2/pages/%s", pageID), "", bytes.NewReader(payload))
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("unable to marshal response: %s", err)
	}

	return response, err
}

// UpdatePage updates a page, incrementing the version manually
func (client *Client) UpdatePageWithVersionBump(pageID string, createPageBody *CreatePageBody) (Page, error) {
	var response Page

	getPagesResponse, err := client.GetPages(&GetPageseQueryParameters{
		ID:         []string{pageID},
		SpaceID:    []string{createPageBody.SpaceID},
		BodyFormat: "storage",
	})

	if err != nil {
		return response, err
	}

	if len(getPagesResponse.Results) == 0 {
		return response, fmt.Errorf("error UpdatePageWithVersionBump: page not found")
	}

	if len(getPagesResponse.Results) > 1 {
		return response, fmt.Errorf("error UpdatePageWithVersionBump: filter returned more than 1 record")
	}

	createPageBody.Version.Number = getPagesResponse.Results[0].Version.Number + 1

	response, err = client.UpdatePage(pageID, createPageBody)
	if err != nil {
		return response, err
	}

	return response, err
}

// UpdatePage update a page by id
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-get
func (client *Client) GetPages(qp *GetPageseQueryParameters) (GetPagesResponse, error) {
	var response GetPagesResponse

	v, _ := query.Values(qp)
	queryParams := v.Encode()

	body, err := client.request("GET", "/api/v2/pages", queryParams, nil)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("unable to marshal response: %s", err)
	}

	return response, err
}

// UpdatePage update a page by id
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-id-delete
func (client *Client) DeletePage(pageID string, qp *DeletePageQueryParameters) error {

	v, _ := query.Values(qp)
	queryParams := v.Encode()

	_, err := client.request("DELETE", fmt.Sprintf("/api/v2/pages/%s", pageID), queryParams, nil)
	if err != nil {
		return err
	}

	return err
}

// UpdatePage update a page by id
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-id-delete
func (client *Client) DeleteAllPagesInSpace(spaceID string) error {

	getPagesInSpaceResponse, err := client.GetPagesInSpace(spaceID, &GetPagesInSpaceQueryParameters{
		Status: []string{"current"},
	})
	if err != nil {
		return err
	}

	for _, page := range getPagesInSpaceResponse.Results {
		err = client.DeletePage(page.ID, &DeletePageQueryParameters{Purge: false})
		if err != nil {
			return err
		}
	}

	return nil
}

type DeletePageQueryParameters struct {
	Purge bool `url:"purge,omitempty"` // If attempting to purge the page.
}

type GetPageseQueryParameters struct {
	ID         []string `url:"id,omitempty"`          // Filter the results based on page ids. Multiple page ids can be specified as a comma-separated list.
	SpaceID    []string `url:"space-id,omitempty"`    // Filter the results based on space ids. Multiple space ids can be specified as a comma-separated list.
	Status     []string `url:"status,omitempty"`      // Filter the results to pages based on their status. By default, current and archived are used.
	Title      string   `url:"title,omitempty"`       // Filter the results to pages based on their title.
	BodyFormat string   `url:"body-format,omitempty"` // The content format types to be returned in the body field of the response. If available, the representation will be available under a response field of the same name under the body field.
	GetResultsQueryParameters
}

type GetResultsQueryParameters struct {
	Cursor string `url:"cursor,omitempty"` // Used for pagination, this opaque cursor will be returned in the next URL in the Link response header. Use the relative URL in the Link header to retrieve the next set of results.
	Limit  int    `url:"limit,omitempty"`  // Maximum number of pages per result to return. If more results exist, use the Link header to retrieve a relative URL that will return the next set of results.
	Sort   string `url:"sort,omitempty"`   // Used to sort the result by a particular field.
}
