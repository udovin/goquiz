package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/labstack/gommon/log"
	"github.com/udovin/solve/config"
)

type LogLevel = config.LogLevel

// Config stores configuration for GoQuiz.
type Config struct {
	// DB contains database connection config.
	DB DB `json:"db"`
	// SocketFile contains path to socket.
	SocketFile string `json:"socket_file"`
	// Server contains API server config.
	Server *Server `json:"server"`
	// Security contains security config.
	Security *Security `json:"security"`
	// LogLevel contains level of logging.
	//
	// You can use following values:
	//  * debug
	//  * info (default)
	//  * warn
	//  * error
	//  * off
	LogLevel LogLevel `json:"log_level,omitempty"`
}

// Server contains server config.
type Server struct {
	// Host contains server host.
	Host string `json:"host"`
	// Port contains server port.
	Port int `json:"port"`
	// Static contains path to static files.
	Static string `json:"static"`
}

// Address returns string representation of server address.
func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Security contains security config.
type Security struct {
	// PasswordSalt contains salt for password hashing.
	PasswordSalt string `json:"password_salt"`
}

var configFuncs = template.FuncMap{
	"json": func(value interface{}) (string, error) {
		data, err := json.Marshal(value)
		return string(data), err
	},
	"file": func(name string) (string, error) {
		bytes, err := ioutil.ReadFile(name)
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(bytes), "\r\n"), nil
	},
	"env": os.Getenv,
}

// LoadFromFile loads configuration from json file.
func LoadFromFile(file string) (Config, error) {
	cfg := Config{
		SocketFile: "/tmp/goquiz-server.sock",
		// By default we should use INFO level.
		LogLevel: LogLevel(log.INFO),
	}
	tmpl, err := template.New(filepath.Base(file)).
		Funcs(configFuncs).ParseFiles(file)
	if err != nil {
		return Config{}, err
	}
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, nil); err != nil {
		return Config{}, err
	}
	if err := json.NewDecoder(&buffer).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
