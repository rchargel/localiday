package server

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v2"
)

var configuration *Application

// LoadConfiguration loads the configuration singleton.
func LoadConfiguration() *Application {
	if configuration != nil {
		return configuration
	}
	configuration = loadConfiguration()
	return configuration
}

// Application describes the current status and version of the application.
type Application struct {
	Name        string
	Description string
	DBVersion   uint16
	Version     string
}

// ToString prints out a string representation of the configuration.
func (a *Application) ToString() string {
	return a.Name + " version: " + fmt.Sprintf("%v", a.Version) + "\n" + a.Description
}

func loadConfiguration() *Application {
	data, err := ioutil.ReadFile("conf/application.yaml")
	if err != nil {
		panic(err)
	}
	m := make(map[string]string)
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		panic(err)
	}
	v, err := strconv.ParseUint(m["DBVersion"], 10, 16)
	if err != nil {
		panic(err)
	}

	return &Application{Name: m["Name"], Description: m["Description"], DBVersion: uint16(v), Version: m["Version"]}
}
