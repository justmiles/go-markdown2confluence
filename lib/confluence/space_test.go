package confluence

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

var (
	testSpace1         Space
	getSpacesResponse1 string
)

func init() {
	date, _ := time.Parse("2006-01-02", "2001-01-01")

	testSpace1 = Space{
		Name:        "TEST",
		CreatedAt:   date,
		AuthorID:    "000000:000000-0000-0000-0000-000000000000",
		HomepageID:  "0",
		Icon:        nil,
		Key:         "TEST",
		ID:          "000000",
		Type:        "global",
		Description: nil,
		Status:      "current",
		Links: Links{
			Webui: "/spaces/TEST",
		},
	}

	testSpace1Json, _ := testSpace1.json()
	getSpacesResponse1 = fmt.Sprintf(`{"results":[%s],"_links":{}}`, testSpace1Json)

}

func TestClient_GetSpaces(t *testing.T) {

	gock.New("http://localhost").
		Get("/wiki/api/v2/spaces").
		Reply(200).
		BodyString(getSpacesResponse1)

	type args struct {
		qp *GetSpacesQueryParameters
	}
	tests := []struct {
		name    string
		args    args
		want    GetSpacesResponse
		wantErr bool
	}{
		{
			name:    "happy path",
			wantErr: false,
			args: args{
				qp: &GetSpacesQueryParameters{
					Keys: []string{"TEST"},
				},
			},
			want: GetSpacesResponse{
				Results: []Space{testSpace1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTestClient()

			got, err := client.GetSpaces(tt.args.qp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetSpaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetSpaces() = %v, want %v", got, tt.want)
			}
		})
	}

	assert.Equal(t, gock.IsDone(), true, "all gock routes should be exercised")

}
