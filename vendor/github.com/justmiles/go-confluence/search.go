package confluence

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/go-querystring/query"
	log "github.com/sirupsen/logrus"
)

// Search searches for content using the Confluence Query Language (CQL)
// https://developer.atlassian.com/cloud/confluence/rest/#api-search-get
//
// Example:
//   searchResults, err := client.Search(&confluence.SearchQueryParameters{
//     CQL:   "space = PE",
//     Limit: 1,
//   })
//
//   if err != nil {
//     errorAndExit(err)
//   }
//
//   for _, searchResult := range searchResults {
//     fmt.Println(searchResult.Title)
//   }
func (client *Client) Search(qp *SearchQueryParameters) ([]SearchResult, error) {
	var queryParams string
	if qp != nil {
		v, _ := query.Values(qp)
		queryParams = v.Encode()
	}
	var searchResponse SearchResponse

	body, err := client.request("GET", "/rest/api/search", queryParams, "")
	if err != nil {
		return searchResponse.Results, err
	}
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		log.Error("Unable to unmarshal SearchResponse. Received: '", string(body), "'")
	}

	if searchResponse.Message != "" {
		err = errors.New(searchResponse.Message)
	}
	return searchResponse.Results, err

}

// SearchQueryParameters query parameters for Search
type SearchQueryParameters struct {
	CQL                   string `url:"cql"`
	CQLContext            string `url:"cqlcontext,omitempty"`
	IncludeArchivedSpaces bool   `url:"includeArchivedSpaces,omitempty"`
	Limit                 int    `url:"limit,omitempty"`
	Start                 int    `url:"start,omitempty"`
}

// SearchResponse represents the data returned from the Confluence API
type SearchResponse struct {
	APIResponse
	Results        []SearchResult `json:"results,omitempty"`
	Start          int            `json:"start,omitempty"`
	Limit          int            `json:"limit,omitempty"`
	Size           int            `json:"size,omitempty"`
	TotalSize      int            `json:"totalSize,omitempty"`
	CqlQuery       string         `json:"cqlQuery,omitempty"`
	SearchDuration int            `json:"searchDuration,omitempty"`
	Links          struct {
		Base    string `json:"base,omitempty"`
		Context string `json:"context,omitempty"`
	} `json:"_links,omitempty"`
}

// SearchResult results from Search
type SearchResult struct {
	Space struct {
		Key      string `json:"key,omitempty"`
		Name     string `json:"name,omitempty"`
		Type     string `json:"type,omitempty"`
		Metadata struct {
		} `json:"metadata,omitempty"`
		Status     string `json:"status,omitempty"`
		Expandable struct {
			Operations  string `json:"operations,omitempty"`
			Permissions string `json:"permissions,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"_expandable,omitempty"`
		Links struct {
			Self string `json:"self,omitempty"`
		} `json:"_links,omitempty"`
	} `json:"space,omitempty"`
	Title                 string `json:"title,omitempty"`
	Excerpt               string `json:"excerpt,omitempty"`
	URL                   string `json:"url,omitempty"`
	ResultGlobalContainer struct {
		Title      string `json:"title"`
		DisplayURL string `json:"displayUrl"`
	} `json:"resultGlobalContainer"`
	Breadcrumbs          []interface{} `json:"breadcrumbs,omitempty"`
	EntityType           string        `json:"entityType,omitempty"`
	IconCSSClass         string        `json:"iconCssClass,omitempty"`
	LastModified         time.Time     `json:"lastModified,omitempty"`
	FriendlyLastModified string        `json:"friendlyLastModified,omitempty"`
	Score                float64       `json:"score,omitempty"`
	Content              `json:"content,omitempty"`
}
