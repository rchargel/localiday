package util

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"time"

	"github.com/hoisie/web"
)

// Constants for writing to output to the browser.
const (
	HTTPAcceptEncoding  = "Accept-encoding"
	HTTPContentEncoding = "Content-encoding"
	HTTPContentLength   = "Content-length"
	HTTPLastModified    = "Last-modified"
	HTTPIfModifiedSince = "If-modified-since"

	HTTPBadRequestCode    = 400
	HTTPUnauthorizedCode  = 401
	HTTPForbiddenCode     = 403
	HTTPFileNotFoundCode  = 404
	HTTPInvalidMethodCode = 405
	HTTPServerErrorCode   = 500

	dateFormat = "Mon 2 Jan 2006 15:04:05 MST"
)

// NewResponseWriter creates a new response writer.
func NewResponseWriter(ctx *web.Context) *ResponseWriter {
	return &ResponseWriter{"text/html", -1, make(map[string]string, 5), ctx}
}

// ResponseWriter used to write content back to the browser.
type ResponseWriter struct {
	Format       string
	LastModified int64
	Headers      map[string]string
	*web.Context
}

// Closable an interface which defines types that can be closed.
type Closable interface {
	Close() error
}

// IsModified checks to see if the content has been modified. If not, it will
// output a 304 response.
func (w *ResponseWriter) IsModified() bool {
	if w.LastModified > 0 {
		header := w.Request.Header.Get(HTTPIfModifiedSince)
		lastMod := time.Unix(w.LastModified, 0).UTC()
		if requestDate, err := time.ParseInLocation(dateFormat, header, time.UTC); err == nil {
			if lastMod.After(requestDate) || requestDate == lastMod {
				w.NotModified()
				return false
			}
		}
	}
	return true
}

// SendError sends an error message and error code back to the browser.
func (w *ResponseWriter) SendError(code int, err error) {
	w.Abort(code, err.Error())
}

// SendFile outputs a file to the browser including content type.
func (w *ResponseWriter) SendFile(contentType, filepath string) {
	w.Format = contentType
	if file, err := os.Open(filepath); err != nil {
		w.SendError(HTTPFileNotFoundCode, err)
	} else {
		if fileInfo, err := file.Stat(); err != nil {
			w.SendError(HTTPServerErrorCode, err)
		} else {
			length := fileInfo.Size()
			w.Headers[HTTPContentLength] = fmt.Sprint(length)

			w.Respond(file)
		}
	}
}

// SendJSON sends JSON output to the browser.
func (w *ResponseWriter) SendJSON(data interface{}) {
	w.Format = "application/json"
	w.sendHeaders()
	enc := json.NewEncoder(w.ResponseWriter)
	enc.Encode(data)
}

// SendTemplate sends the output of the template to the browser.
func (w *ResponseWriter) SendTemplate(t *template.Template, data interface{}) {
	w.sendHeaders()
	t.Execute(w.ResponseWriter, data)
}

// Respond sends the response to the browser.
func (w *ResponseWriter) Respond(reader io.Reader) {
	w.sendHeaders()
	w.compressOutputFilter(reader)
}

func (w *ResponseWriter) sendHeaders() {
	w.ContentType(w.getContentType())

	if len(w.Headers) > 0 {
		for header, value := range w.Headers {
			w.Header().Add(header, value)
		}
	}

	if w.LastModified >= 0 {
		date := time.Unix(w.LastModified, 0)
		formattedDate := date.UTC().Format(dateFormat)

		w.Header().Add(HTTPLastModified, formattedDate)
	}
}

func (w *ResponseWriter) isCompressable() bool {
	return w.Format != "text/html" && !strings.Contains(w.Format, "image/")
}

func (w *ResponseWriter) getContentType() string {
	if strings.Contains(w.Format, "image/") {
		return w.Format
	}
	return w.Format + "; charset=utf-8"
}

func (w *ResponseWriter) compressOutputFilter(reader io.Reader) {
	header := w.Request.Header
	acceptEncoding := header.Get(HTTPAcceptEncoding)
	if strings.Index(acceptEncoding, "gzip") != -1 && w.isCompressable() {
		w.Header().Add(HTTPContentEncoding, "gzip")
		gzw := gzip.NewWriter(w.ResponseWriter)
		writer := bufio.NewWriter(gzw)
		io.Copy(writer, reader)
		writer.Flush()
		gzw.Flush()
		gzw.Close()
	} else {
		io.Copy(w.ResponseWriter, reader)
	}
	if c, isCloseable := reader.(Closable); isCloseable {
		c.Close()
	}
}
