// Package auth manages the secrets required to use the API.
package auth

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

// Config is the go representation of the YAML file containing the secrets.
type Config struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

// Load returns the username, password, stored in the config file.
func Load(file string) (string, string, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return "", "", fmt.Errorf("ReadFile(%q): %w", file, err)
	}
	var c Config
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return "", "", fmt.Errorf("unmarshal %q to yaml: %w", file, err)
	}
	return c.User, c.Pass, nil
}

// MustLoad loads the username and password and exits with an error if it can't.
func MustLoad(file string) (string, string) {
	u, p, err := Load(file)
	if err != nil {
		glog.Exitf("auth.MustLoad(%q): %v", file, err)
	}

	if u == "" {
		glog.Exitf("user not specified in %q", file)
	}
	if p == "" {
		glog.Exitf("pass not specified in %q", file)
	}
	return u, p
}
