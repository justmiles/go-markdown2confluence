package lib

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	renderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"

	ce "github.com/justmiles/go-markdown2confluence/lib/extension"
)

func (m *Markdown2Confluence) renderContent(filePath, s string) (string, []string, error) {

	files := make(map[string]string)
	for _, f := range m.files {
		if f.RemoteID != "" {
			files[f.Path] = fmt.Sprintf("%s%s", m.Endpoint, f.Link)
		}
	}

	confluenceExtension := ce.NewConfluenceExtension(filePath, files)

	renderOptions := []renderer.Option{
		html.WithXHTML(),
	}

	if m.WithHardWraps {
		renderOptions = append(renderOptions, html.WithHardWraps())
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.DefinitionList),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(renderOptions...),
		goldmark.WithExtensions(
			confluenceExtension,
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		return "", nil, err
	}

	return buf.String(), confluenceExtension.Images(), nil
}
