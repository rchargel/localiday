package web

import (
	"html/template"
	"io/ioutil"
	"os"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/app"
)

const (
	htmlDir = "assets/html"
)

// HTMLController controller for static HTML documents.
type HTMLController struct {
	lastModified map[string]int64
	templates    map[string]*template.Template
}

// CreateHTMLController creates a new HTML controller.
func CreateHTMLController() *HTMLController {
	return &HTMLController{make(map[string]int64, 10), make(map[string]*template.Template, 10)}
}

// Render renders the html file.
func (c *HTMLController) Render(ctx *web.Context, file string) {
	w := NewResponseWriter(ctx)
	lm := c.getLastModified(file)
	if lm > 0 {
		w.LastModified = lm
	}

	if w.IsModified() {
		f, err := os.Open(htmlDir + "/" + file)
		defer f.Close()

		if err != nil {
			w.SendError(HTTPFileNotFoundCode, err)
		} else {
			w.Format = "text/html"
			lm = time.Now().Unix()
			c.lastModified[file] = lm
			w.LastModified = lm
			w.Respond(f)
		}
	}
}

// RenderRoot renders the root index file.
func (c *HTMLController) RenderRoot(ctx *web.Context, data string) {
	w := NewResponseWriter(ctx)
	lm := c.getLastModified("/")
	if lm > 0 {
		w.LastModified = lm
	}

	if w.IsModified() {
		if t, err := c.getTemplate("/", "localiday.html"); err != nil {
			w.SendError(HTTPServerErrorCode, err)
		} else {
			w.Format = "text/html"
			lm = time.Now().Unix()
			w.LastModified = lm
			c.lastModified["/"] = lm
			w.SendTemplate(t, app.LoadConfiguration())
		}
	}
}

func (c *HTMLController) getTemplate(name, path string) (*template.Template, error) {
	var err error
	if t, found := c.templates[name]; found {
		return t, nil
	}
	if data, err := ioutil.ReadFile(htmlDir + "/" + path); err == nil {
		return template.New(name).Parse(string(data))
	}
	return nil, err
}

func (c *HTMLController) getLastModified(path string) int64 {
	if lastMod, found := c.lastModified[path]; found {
		return lastMod
	}
	return -1
}
