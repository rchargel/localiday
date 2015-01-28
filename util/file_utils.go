package util

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/dchest/cssmin"
	"github.com/rchargel/localiday/conf"
)

// Log levels
const (
	Debug = "DEBUG"
	Info  = "INFO"
	Warn  = "WARN"
	Error = "ERROR"
	Fatal = "FATAL"
)

var logConfig *conf.Application

// CSSMinifier a CSS Minifier Reader.
type CSSMinifier struct {
	*bytes.Reader
}

// NewCSSMinifier creates a new css minifier.
func NewCSSMinifier(data []byte) *CSSMinifier {
	mindata := cssmin.Minify(data)
	return &CSSMinifier{bytes.NewReader(mindata)}
}

// MakeFirstLetterUpperCase makes the first letter uppercase.
func MakeFirstLetterUpperCase(s string) string {

	if len(s) < 2 {
		return strings.ToUpper(s)
	}

	bts := []byte(s)

	lc := bytes.ToUpper([]byte{bts[0]})
	rest := bts[1:]

	return string(bytes.Join([][]byte{lc, rest}, nil))
}

// Log logs output given a log level.
func Log(logLevel, message string, args ...interface{}) {
	if canWriteLog(logLevel) {
		message = fmt.Sprintf("%6s: ", logLevel) + message
		if logLevel == Fatal {
			log.Fatalln(fmt.Sprintf(message, args...))
		} else {
			log.Println(fmt.Sprintf(message, args...))
		}
	}
}

// Contains checks to see if a slice contains a given value.
func Contains(slice []string, value string) bool {
	for _, a := range slice {
		if a == value {
			return true
		}
	}
	return false
}

func canWriteLog(logLevel string) bool {
	ll := getLogLevel()
	switch logLevel {
	case Debug:
		return ll == Debug
	case Info:
		return ll == Debug || ll == Info
	case Warn:
		return ll == Debug || ll == Info || ll == Warn
	case Error:
		return ll == Debug || ll == Info || ll == Warn || ll == Error
	}
	return true
}

func getLogLevel() string {
	if logConfig == nil {
		logConfig = conf.LoadConfiguration()
	}
	return logConfig.LogLevel
}
