package app

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dchest/cssmin"
)

// Log levels
const (
	Debug = "DEBUG"
	Info  = "INFO"
	Warn  = "WARN"
	Error = "ERROR"
	Fatal = "FATAL"
)

var logConfig *Application

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

// ToStringSlice converts an interface slice into a string slice.
func ToStringSlice(slice []interface{}) []string {
	s := make([]string, len(slice))
	for i, v := range slice {
		s[i] = v.(string)
	}
	return s
}

// ToStringValue converts an interface into its string value.
func ToStringValue(n interface{}) string {
	switch n.(type) {
	default:
		return fmt.Sprintf("%v", n)
	case float64:
		return strconv.FormatFloat(n.(float64), 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(n.(float32)), 'f', -1, 64)
	case int64:
		return strconv.Itoa(int(n.(int64)))
	case int32:
		return strconv.Itoa(int(n.(int32)))
	case int:
		return strconv.Itoa(n.(int))
	}
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
		logConfig = LoadConfiguration()
	}
	return logConfig.LogLevel
}
