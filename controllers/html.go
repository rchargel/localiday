package controllers

import (
	"html/template"
	"io/ioutil"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/conf"
	"github.com/rchargel/localiday/util"
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

// RenderRoot renders the root index file.
func (c *HTMLController) RenderRoot(ctx *web.Context) {
	w := util.NewResponseWriter(ctx)
	lm := c.getLastModified("/")
	if lm > 0 {
		w.LastModified = lm
	}

	if w.IsModified() {
		if t, err := c.getTemplate("/", "localiday.html"); err != nil {
			w.SendError(util.HTTPServerErrorCode, err)
		} else {
			w.Format = "text/html"
			lm = time.Now().Unix()
			w.LastModified = lm
			c.lastModified["/"] = lm
			w.SendTemplate(t, conf.LoadConfiguration())
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
