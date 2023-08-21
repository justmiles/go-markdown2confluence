package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"

	r "github.com/justmiles/go-markdown2confluence/lib/renderer"
)

// Confluence is a Goldmark extension that renders markdown content compatable with Confluence
type Confluence struct {
	imageHTMLRender  *r.ConfluenceImageHTMLRender
	metadataRenderer *r.MetadataRenderer
}

// NewConfluenceExtension returns an instanciated instance of Confluence
func NewConfluenceExtension(filePath string) *Confluence {
	c := &Confluence{
		imageHTMLRender:  r.NewConfluenceImageHTMLRender(filePath),
		metadataRenderer: r.NewMetadataRenderer(),
	}
	return c
}

// Images returns a slice of image paths for later upload
func (c *Confluence) Images() []string {
	return c.imageHTMLRender.Images
}

// Extend markdown custom HTML render
func (c *Confluence) Extend(m goldmark.Markdown) {

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(r.NewConfluenceFencedCodeBlockHTMLRender(), 100),
		util.Prioritized(r.NewConfluenceCodeBlockHTMLRender(), 100),
		util.Prioritized(c.metadataRenderer, 100),
		util.Prioritized(c.imageHTMLRender, 100),
	))

}
