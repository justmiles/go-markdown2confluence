package renderer

import (
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// MetadataRenderer is a custom renderer for extracting metadata
type MetadataRenderer struct {
	Title string
	Tags  []string
}

func NewMetadataRenderer() *MetadataRenderer {
	return &MetadataRenderer{}
}

func (r *MetadataRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindDocument, r.renderDocument)
}

func (r *MetadataRenderer) renderDocument(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {

		// Create a Markdown parser with the meta extension
		md := goldmark.New(goldmark.WithExtensions(meta.Meta))

		// Parse the Markdown content
		pctx := parser.NewContext()
		md.Parser().Parse(text.NewReader(source), parser.WithContext(pctx))

		// Retrieve the metadata from the parsed context
		if metadata := meta.Get(pctx); metadata != nil {
			if title, ok := metadata["title"].(string); ok {
				r.Title = title
			}
			if tags, ok := metadata["tags"].([]interface{}); ok {
				for _, tag := range tags {
					if tagStr, ok := tag.(string); ok {
						r.Tags = append(r.Tags, tagStr)
					}
				}
			}
		}

		fmt.Println("title")
		fmt.Println(r.Title)
		fmt.Println("tags")
		fmt.Println((r.Tags))
	}

	// Continue rendering other nodes
	return ast.WalkContinue, nil
}
