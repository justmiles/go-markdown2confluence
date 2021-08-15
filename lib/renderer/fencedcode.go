package renderer

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// ConfluenceFencedCodeBlockHTMLRender is a renderer.NodeRenderer implementation that
// renders FencedCodeBlock nodes.
type ConfluenceFencedCodeBlockHTMLRender struct {
	html.Config
}

// NewConfluenceFencedCodeBlockHTMLRender returns a new ConfluenceFencedCodeBlockHTMLRender.
func NewConfluenceFencedCodeBlockHTMLRender(opts ...html.Option) renderer.NodeRenderer {
	r := &ConfluenceFencedCodeBlockHTMLRender{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *ConfluenceFencedCodeBlockHTMLRender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderConfluenceFencedCode)
}

func (r *ConfluenceFencedCodeBlockHTMLRender) renderConfluenceFencedCode(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	language := n.Language(source)
	// Initialize the language string with an ampty string
	// for easier comparisson later
	langString := ""
	if language != nil {
		langString = string(language)
	}
	if entering {
		// If it is a macro create the macro
		if langString == "CONFLUENCE-MACRO" {
			r.writeMacro(w, source, n)
		} else {
			// else insert a code-macro
			s := `<ac:structured-macro ac:name="code" ac:schema-version="1">`
			s = s + `<ac:parameter ac:name="theme">Confluence</ac:parameter>`
			s = s + `<ac:parameter ac:name="linenumbers">true</ac:parameter>`

			if language != nil {
				s = s + `<ac:parameter ac:name="language">` + langString + `</ac:parameter>`
			}

			s = s + `<ac:plain-text-body><![CDATA[ `
			_, _ = w.WriteString(s)
			r.writeLines(w, source, n)
		}
	} else if langString != "CONFLUENCE-MACRO" {
		// No special handling for the CONFLUENCE-MACRO, just for the code macros
		s := ` ]]></ac:plain-text-body></ac:structured-macro>`
		_, _ = w.WriteString(s)
	}
	return ast.WalkContinue, nil
}

func (r *ConfluenceFencedCodeBlockHTMLRender) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		w.WriteString(string(line.Value(source)))
	}
}

func (r *ConfluenceFencedCodeBlockHTMLRender) writeMacro(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	// prepare the macrostart
	macrostart := strings.Builder{}
	macrostart.WriteString(`<ac:structured-macro`)
	// and initialize the parameters
	parameters := strings.Builder{}
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		text := string(line.Value(source))
		// Split the line at the first colon
		keyValue := strings.SplitN(text, ":", 2)
		// Ignore lines which didn't split into two parts
		if len(keyValue) == 2 {
			// key is left of the colon
			key := strings.TrimSpace(keyValue[0])
			// value is to the right. We trim both
			value := strings.TrimSpace(keyValue[1])
			// If the key was not indented
			if key[0] == keyValue[0][0] {
				// we append a new attribute to the macro
				macrostart.WriteString(` ac:` + key + `="` + value + `"`)
			} else {
				// It is aparameter to the macro
				parameters.WriteString(`<ac:parameter ac:name="` + key + `">` + value + `</ac:parameter>`)
			}
		} else if len(keyValue) == 1 {
			value := strings.TrimSpace(keyValue[0])
			// assume the name of the param is empty
			parameters.WriteString(`<ac:parameter ac:name="">` + value + `</ac:parameter>`)
		}
	}
	// write the macro start
	w.WriteString(macrostart.String())
	w.WriteString(">")
	// and all parameters
	w.WriteString(parameters.String())
	// and finish it off
	w.WriteString("</ac:structured-macro>")
}
