package lib

import (
	"testing"
)

func TestMarkdownFile_Upload(t *testing.T) {
	t.Skip()
	type args struct {
		m *Markdown2Confluence
	}
	testInstance := Markdown2Confluence{
		Space: "TEST",
	}

	err := testInstance.Init()
	if err != nil {
		t.Error(err)
	}

	defer testInstance.Close()

	tests := []struct {
		name        string
		fields      MarkdownFile
		args        args
		wantUrlPath string
		wantErr     bool
	}{
		{
			name: "happy path",
			args: args{m: &testInstance},
			fields: MarkdownFile{
				ID:     "/about/contributing.md",
				Path:   "mkdocs-1.5.3/docs/about/contributing.md",
				Title:  "Contributing",
				Parent: "/",
				Status: "CREATE",
				MD5Sum: "fcb718fe7d9e253ba3527caee7e48c08",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := &MarkdownFile{
				ID:     tt.fields.ID,
				Path:   tt.fields.Path,
				Title:  tt.fields.Title,
				Parent: tt.fields.Parent,
				Status: tt.fields.Status,
				MD5Sum: tt.fields.MD5Sum,
			}
			_, err := mf.Upload(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkdownFile.Upload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
