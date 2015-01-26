package controllers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/dchest/jsmin"
	"github.com/hoisie/web"
	"github.com/rchargel/localiday/util"
	"gopkg.in/yaml.v2"
)

const (
	jsDir      = "assets/js"
	mainJSFile = "localiday"
	jsFormat   = "application/javascript"
)

// JSController conroller for providing javascript files.
type JSController struct {
	lastModified map[string]int64
}

type assets struct {
	assetFiles []string
}

// CreateJSController creates an instance of the Javascript controller.
func CreateJSController() *JSController {
	return &JSController{make(map[string]int64, 10)}
}

// RenderJS renders the javascript files.
func (c *JSController) RenderJS(ctx *web.Context, version string) {
	w := util.NewResponseWriter(ctx)
	lm := c.getLastModified(mainJSFile)
	if lm > 0 {
		w.LastModified = lm
	}

	if w.IsModified() {
		if reader, err := readJS(); err != nil {
			w.SendError(util.HTTPServerErrorCode, err)
		} else {
			w.Format = jsFormat
			lm = time.Now().Unix()
			c.lastModified[mainJSFile] = lm
			w.LastModified = lm
			w.Respond(reader)
		}
	}
}

// RenderJSFile renders a specific javascript file.
func (c *JSController) RenderJSFile(ctx *web.Context, file string) {
	w := util.NewResponseWriter(ctx)
	lm := c.getLastModified(file)
	if lm > 0 {
		w.LastModified = lm
	}
	if w.IsModified() {
		if data, err := ioutil.ReadFile(jsDir + "/" + file); err != nil {
			w.SendError(util.HTTPFileNotFoundCode, err)
		} else {
			if mdata, err := jsmin.Minify(data); err == nil {
				w.Format = jsFormat
				lm = time.Now().Unix()
				c.lastModified[file] = lm
				w.LastModified = lm
				w.Headers[util.HTTPContentLength] = fmt.Sprint(len(mdata))
				w.Respond(bytes.NewReader(mdata))
			} else {
				util.Log(util.Error, "Could not minimize data: %v", err)
				w.SendError(util.HTTPServerErrorCode, err)
			}
		}
	}
}

func readJS() (io.Reader, error) {
	var reader io.Reader
	var err error
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	if a, err := loadAssets(); err == nil {
		for _, asset := range a.assetFiles {
			if err = readJSIntoFile(asset, writer); err != nil {
				return reader, err
			}
			writer.Flush()
		}
		reader = bytes.NewReader(buffer.Bytes())
	}
	return reader, err
}

func loadAssets() (assets, error) {
	a := assets{}
	assetData, err := ioutil.ReadFile(jsDir + "/jsresources.yaml")
	if err != nil {
		return a, err
	}

	m := make(map[string][]string)

	if err = yaml.Unmarshal(assetData, &m); err == nil {
		a.assetFiles = m["resources"]
	}

	return a, err
}

func readJSIntoFile(asset string, writer io.Writer) error {
	var err error
	file := jsDir + "/" + asset
	if script, err := ioutil.ReadFile(file); err == nil {
		/* // UNCOMMENT FOR MINIFICATION
		if strings.Contains(asset, ".min.") {
			writer.Write(script)
		} else {
			if c, err := jsmin.Minify(script); err == nil {
				writer.Write(c)
			} else {
				return err
			}
		}
		*/
		writer.Write(script)
	}
	return err
}

func (c *JSController) getLastModified(file string) int64 {
	if i, found := c.lastModified[file]; found {
		return i
	}
	return -1
}
