package lib

import (
	"testing"
)

func TestMarkdown2Confluence_Validate(t *testing.T) {
	type fields struct {
		Space          string
		Title          string
		SourceMarkdown []string

		Username    string
		Password    string
		Endpoint    string
		AccessToken string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Space:          "test",
				Username:       "test",
				Password:       "test",
				Endpoint:       "https://demo.atlassian.net/wiki",
				Title:          "test",
				SourceMarkdown: []string{"."},
			},
		},
		{
			name:    "invalid url",
			wantErr: true,
			fields: fields{
				Space:          "test",
				Username:       "test",
				Password:       "test",
				Endpoint:       "https://demo.atlassian.n et/wiki",
				Title:          "test",
				SourceMarkdown: []string{"."},
			},
		},
		{
			name:    "space not defined",
			wantErr: true,
			fields: fields{
				Username:       "test",
				Password:       "test",
				Endpoint:       "https://demo.atlassian.n et/wiki",
				Title:          "test",
				SourceMarkdown: []string{"."},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Markdown2Confluence{
				Space:          tt.fields.Space,
				SourceMarkdown: tt.fields.SourceMarkdown,
			}
			m.Username = tt.fields.Username
			m.Password = tt.fields.Password
			m.Endpoint = tt.fields.Endpoint
			m.AccessToken = tt.fields.AccessToken
			if err := m.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Markdown2Confluence.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
