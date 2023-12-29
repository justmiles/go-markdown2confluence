package confluence

import (
	"io"
	"reflect"
	"testing"
)

func NewTestClient() *Client {
	return &Client{
		Username: "user@email.com",
		Password: "password",
		Endpoint: "http://localhost/wiki",
		LogLevel: "info",
	}
}

func TestClient_request(t *testing.T) {
	type fields struct {
		Cookie      string
		Username    string
		Password    string
		AccessToken string
		Endpoint    string
		LogLevel    string
	}
	type args struct {
		method      string
		apiEndpoint string
		queryParams string
		payload     io.Reader
		preFns      []PreRequestFn
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Cookie:      tt.fields.Cookie,
				Username:    tt.fields.Username,
				Password:    tt.fields.Password,
				AccessToken: tt.fields.AccessToken,
				Endpoint:    tt.fields.Endpoint,
				LogLevel:    "debug",
			}
			got, err := client.request(tt.args.method, tt.args.apiEndpoint, tt.args.queryParams, tt.args.payload, tt.args.preFns...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.request() = %v, want %v", got, tt.want)
			}
		})
	}
}
