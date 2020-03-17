package confluence

import (
	"encoding/json"
	"strings"

	"github.com/google/go-querystring/query"
	log "github.com/sirupsen/logrus"
)

func (client *Client) labelEndpoint(contentID string) string {
	return "/rest/api/content/" + contentID + "/label"
}

// GetContent Returns all content in a Confluence instance.
// https://developer.atlassian.com/cloud/confluence/rest/#api-content-get
func (client *Client) GetContent(qp *GetContentQueryParameters) ([]Content, error) {

	qp.ExpandString = strings.Join(qp.Expand, ",")
	v, _ := query.Values(qp)
	queryParams := v.Encode()

	body, err := client.request("GET", "/rest/api/content", queryParams, "")
	if err != nil {
		return nil, err
	}
	var contentResponse ContentResponse
	err = json.Unmarshal(body, &contentResponse)
	if err != nil {
		log.Error("Unable to unmarshal ContentResponse. Received: '", string(body), "'")
	}
	return contentResponse.Results, err
}

// GetContentQueryParameters query parameters for GetContent
type GetContentQueryParameters struct {
	QueryParameters
	Expand       []string `url:"-"`
	ExpandString string   `url:"expand,omitempty"`
	Limit        int      `url:"limit,omitempty"`
	Orderby      string   `url:"orderby,omitempty"`
	PostingDay   string   `url:"postingDay,omitempty"`
	Spacekey     string   `url:"spaceKey,omitempty"`
	Start        int      `url:"start,omitempty"`
	Title        string   `url:"title,omitempty"`
	Trigger      string   `url:"trigger,omitempty"`
	Type         string   `url:"type,omitempty"`
}

// CreateContent creates a new piece of content or publishes an existing draft.
// https://developer.atlassian.com/cloud/confluence/rest/#api-content-post
func (client *Client) CreateContent(bp *CreateContentBodyParameters, qp *QueryParameters) (Content, error) {
	var res Content
	var queryParams string
	if qp != nil {
		v, _ := query.Values(qp)
		queryParams = v.Encode()
	}

	byteString, err := json.Marshal(bp)
	if err != nil {
		log.Error("Unable to marshal body. Received: '", err, "'")
	}

	body, err := client.request("POST", "/rest/api/content", queryParams, string(byteString))
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Error(body)
		log.Error(err)
		log.Error("Unable to unmarshal CreateContentResponse. Received: '", string(body), "'")
	}
	return res, err
}

// UpdateContent updates a piece of content. Use this method to update the title or body of a piece of content, change the status, change the parent page, and more.
// https://developer.atlassian.com/cloud/confluence/rest/#api-content-id-put
func (client *Client) UpdateContent(content *Content, qp *QueryParameters) (Content, error) {
	var queryParams string
	if qp != nil {
		v, _ := query.Values(qp)
		queryParams = v.Encode()
	}

	byteString, err := json.Marshal(content)
	if err != nil {
		log.Error("Unable to marshal body. Received: '", err, "'")
	}

	body, err := client.request("PUT", "/rest/api/content/"+content.ID, queryParams, string(byteString))
	if err != nil {
		return *content, err
	}
	err = json.Unmarshal(body, &content)
	if err != nil {
		log.Error(body)
		log.Error(err)
		log.Error("Unable to unmarshal UpdateContent response. Received: '", string(body), "'")
	}
	return *content, err
}

// LabelPrefix ...
type LabelPrefix string

const (
	// GlobalPrefix ...
	GlobalPrefix LabelPrefix = "global"
	// LocalPrefix ...
	LocalPrefix LabelPrefix = "local"
)

// AddLabels ...
func (client *Client) AddLabels(contentID string, labels []string, prefix LabelPrefix) error {
	type Label struct {
		Prefix string `json:"prefix"`
		Name   string `json:"name"`
	}
	var labelsContent []Label
	for _, l := range labels {
		labelsContent = append(labelsContent, Label{string(prefix), l})
	}

	jsonbody, err := json.Marshal(labelsContent)
	if err != nil {
		return err
	}
	labelEndpoint := client.labelEndpoint(contentID)
	_, err = client.request("POST", labelEndpoint, "", string(jsonbody))
	if err != nil {
		return err
	}
	return nil
}

// CreateContentBodyParameters query parameters for CreateContent
type CreateContentBodyParameters struct {
	Content
}

// DeleteContent oves a piece of content to the space’s trash or purges it from the trash, depending on the content’s type and status:
//  - If the content’s type is `page` or `blogpost` and its status is `current`, it will be trashed.
//  - If the content’s type is `page` or `blogpost` and its status is `trashed`, the content will be purged from the trash and deleted permanently. Note, you must also set the `status` query parameter to `trashed` in your request.
//  - If the content’s type is `comment` or `attachment`, it will be deleted permanently without being trashed.
// https://developer.atlassian.com/cloud/confluence/rest/#api-content-id-delete
func (client *Client) DeleteContent(content Content) error {
	_, err := client.request("DELETE", "/rest/api/content/"+content.ID, "", "")
	return err
}

// ContentResponse represents the data returned from the Confluence API
type ContentResponse struct {
	Results []Content `json:"results"`
}

// Content represents the data returned from the Confluence API
type Content struct {
	ID        string `json:"id,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	Title     string `json:"title,omitempty"`
	URL       string `json:"url,omitempty"`
	Ancestors []struct {
		ID string `json:"id,omitempty"`
	} `json:"ancestors,omitempty"`
	Space struct {
		Key string `json:"key,omitempty"`
	} `json:"space,omitempty"`
	Version struct {
		Number  int    `json:"number,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"version,omitempty"`
	Body struct {
		Storage struct {
			Value           string        `json:"value,omitempty"`
			Representation  string        `json:"representation,omitempty"`
			EmbeddedContent []interface{} `json:"embeddedContent,omitempty"`
			Expandable      struct {
				Content string `json:"content,omitempty"`
			} `json:"_expandable,omitempty"`
		} `json:"storage,omitempty"`
	} `json:"body,omitempty"`
	Links struct {
		Self   string `json:"self,omitempty"`
		Tinyui string `json:"tinyui,omitempty"`
		Editui string `json:"editui,omitempty"`
		Webui  string `json:"webui,omitempty"`
	} `json:"_links,omitempty"`
}
