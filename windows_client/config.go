package main

import (
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

var config struct {
	URL string `yaml:"url"`
}

func init() {
	// default values
	config.URL = "https://uniclip.lonord.name"

	cDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	dir := path.Join(cDir, "uniclip")
	os.MkdirAll(dir, 0755)
	cFile := path.Join(dir, "config.yml")
	if _, err := os.Stat(cFile); err == nil {
		// config.yml exists
		f, err := os.Open(cFile)
		if err != nil {
			return
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return
		}

		yaml.Unmarshal(b, &config)
	}
}
