package renderer

import (
	"github.com/yuin/goldmark/ast"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// ConfluenceCodeBlockHTMLRender is a renderer.NodeRenderer implementation that
// renders KindCodeBlock nodes.
type ConfluenceCodeBlockHTMLRender struct {
	html.Config
}

// NewConfluenceCodeBlockHTMLRender returns a new ConfluenceCodeBlockHTMLRender.
func NewConfluenceCodeBlockHTMLRender(opts ...html.Option) renderer.NodeRenderer {
	r := &ConfluenceCodeBlockHTMLRender{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}

	r.Config.SetOption("XHTML", true)
	r.Config.SetOption("Unsafe", true)
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *ConfluenceCodeBlockHTMLRender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(gast.KindCodeBlock, r.renderConfluenceCodeBlock)
}

func (r *ConfluenceCodeBlockHTMLRender) renderConfluenceCodeBlock(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		s := `<ac:structured-macro ac:name="code" ac:schema-version="1">`
		s = s + `<ac:parameter ac:name="theme">Confluence</ac:parameter>`
		s = s + `<ac:plain-text-body><![CDATA[`
		_, _ = w.WriteString(s)
		r.writeLines(w, source, n)
	} else {
		s := `]]></ac:plain-text-body></ac:structured-macro>`
		_, _ = w.WriteString(s)
	}
	return ast.WalkContinue, nil
}

func (r *ConfluenceCodeBlockHTMLRender) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		w.WriteString(string(line.Value(source)))
	}
}
