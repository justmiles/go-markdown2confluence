package renderer

import (
	"net/url"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// ConfluenceLinkHTMLRender is a renderer.NodeRenderer implementation that
// renders KindLink nodes.
type ConfluenceLinkHTMLRender struct {
	html.Config
	filePath  string
	files     map[string]string
	Reprocess bool
}

func (r ConfluenceLinkHTMLRender) getLinkByFilename(filename string) string {
	for id, url := range r.files {
		if strings.HasSuffix(filename, id) {
			return url
		}
	}
	return ""
}

// NewConfluenceLinkHTMLRender returns a new ConfluenceLinkHTMLRender.
func NewConfluenceLinkHTMLRender(filePath string, files map[string]string, opts ...html.Option) *ConfluenceLinkHTMLRender {
	r := &ConfluenceLinkHTMLRender{
		Config:   html.NewConfig(),
		filePath: filePath,
		files:    files,
	}

	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}

	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *ConfluenceLinkHTMLRender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.renderConfluenceLink)
}

func (r *ConfluenceLinkHTMLRender) renderConfluenceLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("<a href=\"")
		if r.Unsafe || !html.IsDangerousURL(n.Destination) {

			// If this is a local file and not an HTTP url, then let's render this for Confluence
			parsedURL, _ := url.Parse(string(n.Destination))
			if f, err := localFile(r.filePath, []byte(parsedURL.Path)); err == nil {
				url := r.getLinkByFilename(f)
				// We have a destination that we can't resolve (it hasn't been uploaded yet)
				if url == "" {
					r.Reprocess = true
				} else {
					// we found a valid url, set the destination
					n.Destination = []byte(url)

					// valid url has a fragment. Add that
					if parsedURL.Fragment != "" {
						n.Destination = []byte(url + "#" + parsedURL.Fragment)
					}
				}

			}
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		_ = w.WriteByte('"')
		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			_ = w.WriteByte('"')
		}
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.LinkAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}
