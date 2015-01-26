package controllers

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/util"
	"github.com/yosssi/gcss"
)

const (
	cssDir  = "assets/css"
	gssFile = ".gcss"
	cssFile = ".css"
)

// CSSController a controller used to output css files.
type CSSController struct {
	LastModified int64
}

// CreateCSSController creates an instance of the CSS controller.
func CreateCSSController() *CSSController {
	return &CSSController{LastModified: -1}
}

// RenderCSS renders the CSS output.
func (c *CSSController) RenderCSS(ctx *web.Context, version string) {
	w := util.NewResponseWriter(ctx)
	if c.LastModified > 0 {
		w.LastModified = c.LastModified
	}

	if w.IsModified() {
		if reader, err := readCSS(); err != nil {
			w.SendError(util.HTTPServerErrorCode, err)
		} else {
			w.Format = "text/css"
			c.LastModified = time.Now().Unix()
			w.LastModified = c.LastModified
			w.Respond(reader)
		}
	}
}

func readCSS() (io.Reader, error) {
	var reader io.Reader
	var err error
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	fileOrder := []string{"main", "index", "tablets", "fullscreen"}
	if fileMap, err := readDir(); err == nil {
		for _, filename := range fileOrder {
			readIntoFile(filename, fileMap, writer)
			writer.Flush()
		}
		reader = util.NewCSSMinifier(buffer.Bytes())
	}
	return reader, err
}

func readIntoFile(filename string, fileMap map[string]os.FileInfo, writer io.Writer) error {
	if file, ok := fileMap[filename]; ok {
		if err := readFile(file, fileMap, writer); err != nil {
			return err
		}
	}
	return nil
}

func readFile(fileInfo os.FileInfo, fileMap map[string]os.FileInfo, writer io.Writer) error {
	file, err := os.Open(cssDir + "/" + fileInfo.Name())
	defer file.Close()
	if err != nil {
		return err
	}

	if strings.Contains(fileInfo.Name(), gssFile) {
		util.Log(util.Debug, "Compiling CSS file %v.", file.Name())
		// first grab the mixins
		mixinsInfo := fileMap["mixins"]
		mixins, err := os.Open(cssDir + "/" + mixinsInfo.Name())
		defer mixins.Close()
		if err != nil {
			return err
		}

		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		io.Copy(w, mixins)
		io.Copy(w, file)
		w.Flush()
		c := b.Bytes()
		r := bytes.NewReader(c)
		_, err = gcss.Compile(writer, r)
	} else {
		_, err = io.Copy(writer, file)
	}
	return err
}

func readDir() (map[string]os.FileInfo, error) {
	var m map[string]os.FileInfo
	var err error
	if files, err := ioutil.ReadDir(cssDir); err == nil {
		m = make(map[string]os.FileInfo, len(files))
		for _, file := range files {
			n := file.Name()
			if strings.Contains(n, gssFile) || strings.Contains(n, cssFile) {
				m[n[:strings.LastIndex(n, ".")]] = file
			}
		}
	}
	return m, err
}
