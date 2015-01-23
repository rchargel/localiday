package util

import (
	"bytes"

	"github.com/dchest/cssmin"
)

// CSSMinifier a CSS Minifier Reader.
type CSSMinifier struct {
	*bytes.Reader
}

// NewCSSMinifier creates a new css minifier.
func NewCSSMinifier(data []byte) *CSSMinifier {
	mindata := cssmin.Minify(data)
	return &CSSMinifier{bytes.NewReader(mindata)}
}
