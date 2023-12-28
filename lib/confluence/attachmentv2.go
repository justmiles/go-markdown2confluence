package confluence

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-querystring/query"
)

type GetAttachmentsForPageQueryParameters struct {
	Status    []string `url:"status,omitempty"`
	MediaType string   `url:"mediaType,omitempty"`
	Filename  string   `url:"filename,omitempty"`
	GetResultsQueryParameters
}

type GetAttachmentsForPageResponse struct {
	Results []PageAttachment `json:"results,omitempty"`
	Links   Links            `json:"_links,omitempty"`
}

type PageAttachment struct {
	ID                   string  `json:"id,omitempty"`
	Status               string  `json:"status,omitempty"`
	Title                string  `json:"title,omitempty"`
	CreatedAt            string  `json:"createdAt,omitempty"`
	PageID               string  `json:"pageId,omitempty"`
	BlogPostID           string  `json:"blogPostId,omitempty"`
	CustomContentID      string  `json:"customContentId,omitempty"`
	MediaType            string  `json:"mediaType,omitempty"`
	MediaTypeDescription string  `json:"mediaTypeDescription,omitempty"`
	Comment              string  `json:"comment,omitempty"`
	FileID               string  `json:"fileId,omitempty"`
	FileSize             int     `json:"fileSize,omitempty"`
	WebuiLink            string  `json:"webuiLink,omitempty"`
	DownloadLink         string  `json:"downloadLink,omitempty"`
	Version              Version `json:"version,omitempty"`
	Links                Links   `json:"_links,omitempty"`
}

// GetAttachmentsForPage returns the attachments of specific page
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-attachment/#api-pages-id-attachments-get
func (client *Client) GetPageAttachments(pageID string, qp *GetAttachmentsForPageQueryParameters) (GetAttachmentsForPageResponse, error) {
	var response GetAttachmentsForPageResponse

	v, _ := query.Values(qp)
	queryParams := v.Encode()

	body, err := client.request("GET", fmt.Sprintf("/api/v2/pages/%s/attachments", pageID), queryParams, nil)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("unable to marshal response: %s", err)
	}

	return response, err
}

func (client *Client) GetPageAttachmentByFilename(pageID, filename string) (PageAttachment, error) {
	var response PageAttachment
	getPageAttachments, err := client.GetPageAttachments(pageID, &GetAttachmentsForPageQueryParameters{
		Filename: filename,
	})
	if err != nil {
		return response, err
	}

	for _, pageAttachment := range getPageAttachments.Results {
		return pageAttachment, nil
	}

	return response, fmt.Errorf("attachment not found")
}
