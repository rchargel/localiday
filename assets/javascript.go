package assets

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/dchest/jsmin"
	"github.com/hoisie/web"
	"github.com/rchargel/localiday/util"
	"gopkg.in/yaml.v2"
)

const (
	jsDir = "assets/js"
)

// JSController conroller for providing javascript files.
type JSController struct {
	LastModified int64
}

type assets struct {
	assetFiles []string
}

// CreateJSController creates an instance of the Javascript controller.
func CreateJSController() *JSController {
	return &JSController{LastModified: -1}
}

// RenderJS renders the javascript files.
func (c *JSController) RenderJS(ctx *web.Context, version string) {
	w := util.NewResponseWriter(ctx)
	if c.LastModified > 0 {
		w.LastModified = c.LastModified
	}

	if w.IsModified() {
		if reader, err := readJS(); err != nil {
			w.SendError(util.HTTPServerErrorCode, err)
		} else {
			w.Format = "application/javascript"
			c.LastModified = time.Now().Unix()
			w.LastModified = c.LastModified
			w.Respond(reader)
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
		if output, err := jsmin.Minify(buffer.Bytes()); err == nil {
			reader = bytes.NewReader(output)
		}
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
		writer.Write(script)
	}
	return err
}
