package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	RootDir    string `yaml:"RootDir"`
	LiveReload bool   `yaml:"LiveReload"`
	Dir        string `yaml:"Dir"`
	Graphiql   bool   `yaml:"GraphiQL"`
	Port       string `yaml:"Port"`
}

func ParseYamlFile(path string) (*Conf, error) {
	var c Conf

	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Conf) Default() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	c.RootDir = "/var/graphqld"
	c.LiveReload = false
	c.Dir = home
	c.Graphiql = false
	c.Port = "80"

	return nil
}
