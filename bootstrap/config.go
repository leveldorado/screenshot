package bootstrap

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
}

func readConfig(path string) (config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config{}, fmt.Errorf(`failed to read config file: [path: %s, error: %w]`, path, err)
	}
	var c config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return config{}, fmt.Errorf(`failed to unmarshal config to struct: [data: %s, error: %w]`, data, err)
	}
	return c, nil
}
