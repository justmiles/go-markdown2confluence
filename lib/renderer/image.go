package renderer

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// ConfluenceImageHTMLRender is a renderer.NodeRenderer implementation that
// renders KindImage nodes.
type ConfluenceImageHTMLRender struct {
	html.Config
	Images   []string
	filePath string
}

// NewConfluenceImageHTMLRender returns a new ConfluenceImageHTMLRender.
func NewConfluenceImageHTMLRender(filePath string, opts ...html.Option) *ConfluenceImageHTMLRender {
	r := &ConfluenceImageHTMLRender{
		Config:   html.NewConfig(),
		filePath: filePath,
	}

	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}

	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *ConfluenceImageHTMLRender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderConfluenceImage)
}

func (r *ConfluenceImageHTMLRender) renderConfluenceImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Image)

	// If this is a local file and not an HTTP url, then let's render this for Confluence
	if f, err := localFile(r.filePath, n.Destination); err == nil {
		r.Images = append(r.Images, f)
		_, _ = w.WriteString(`<ac:image><ri:attachment ri:filename="`)
		_, _ = w.WriteString(path.Base(f))
		_, _ = w.WriteString(`"/></ac:image>`)

		return ast.WalkSkipChildren, nil
	}

	// This is a regular HTTP url, render it in normal XHTML
	_, _ = w.WriteString("<img src=\"")
	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	}

	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(n.Text(source))
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		_ = w.WriteByte('"')
	}
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}
	if r.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

// RenderImageAttributes renders an Image's given attributes.
func RenderImageAttributes(w util.BufWriter, node ast.Node, filter util.BytesFilter) {
	for _, attr := range node.Attributes() {
		if filter != nil && !filter.Contains(attr.Name) {
			if !bytes.HasPrefix(attr.Name, []byte("data-")) {
				continue
			}
		}
		_, _ = w.WriteString(" ac:")
		_, _ = w.Write(attr.Name)
		_, _ = w.WriteString(`="`)
		// TODO: convert numeric values to strings
		_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
		_ = w.WriteByte('"')
	}
}

func localFile(filePath string, destination []byte) (string, error) {

	localizedPath := string(destination)
	_, err := os.Stat(localizedPath)
	if err == nil {
		fmt.Println("MY PATH IS "+localizedPath)
		return localizedPath, nil
	}

	//path.Dir currDir is workpath so "path.Dir is '.'"
	//And so make a absolute path for check file
	localizedAbsPath, _ := filepath.Abs(filePath)
	localizedPath = path.Join(filepath.Dir(localizedAbsPath), string(destination))
	_, err = os.Stat(localizedPath)
	if err == nil {
		return localizedPath, nil
	}

	return "", fmt.Errorf("not a local file")
}
