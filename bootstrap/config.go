package bootstrap

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type config struct {
	Queue struct {
		BufferSize           int           `yaml:"buffer_size"`
		ConnectTimeout       time.Duration `yaml:"connect_timeout"`
		HandleMessageTimeout time.Duration `yaml:"handle_message_timeout"`
		WaitReplyTimeout     time.Duration `yaml:"wait_reply_timeout"`
	} `yaml:"queue"`
	Database struct {
		Name        string `yaml:"name"`
		Collections struct {
			Metadata       string `yaml:"metadata"`
			VersionCounter string `yaml:"version_counter"`
		} `yaml:"collections"`
	} `yaml:"database"`
	Screenshot struct {
		Format  string `yaml:"format"`
		Quality int    `yaml:"quality"`
	} `yaml:"screenshot"`
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
