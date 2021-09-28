package main

import (
	_ "embed"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

//go:embed default_config.yml
var defaultCfg []byte

var config struct {
	URL string `yaml:"url"`
}

func init() {
	// default values
	config.URL = "http://127.0.0.1:8080"

	cFile := mustGetConfigFilePath()
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
	} else {
		os.WriteFile(cFile, defaultCfg, 0644)
	}
}

func mustGetConfigFilePath() string {
	cDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	dir := path.Join(cDir, "uniclip")
	os.MkdirAll(dir, 0755)
	return path.Join(dir, "config.yml")
}
