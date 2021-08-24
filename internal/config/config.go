package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	RootDir   string `yaml:"RootDir"`
	HotReload bool   `yaml:"HotReload"`
	Dir       string `yaml:"Dir"`
	Graphiql  bool   `yaml:"GraphiQL"`
	Port      string `yaml:"Port"`
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

func ParseFromEnv() *Conf {
	var (
		c = Conf{
			RootDir: os.Getenv("GRAPHQLD_ROOT_DIR"),
			Dir:     os.Getenv("GRAPHQLD_DIR"),
			Port:    os.Getenv("GRAPHQLD_PORT"),
		}

		defConf = Conf{}
	)

	switch x := os.Getenv("GRAPHQLD_HOT_RELOAD"); x {
	case "":
		c.HotReload = false
	default:
		c.HotReload = true
	}

	switch x := os.Getenv("GRAPHQLD_GRAPHIQL"); x {
	case "":
		c.Graphiql = false
	default:
		c.Graphiql = true
	}

	defConf.Default()
	if c.RootDir == "" {
		c.RootDir = defConf.RootDir
	}

	if c.Dir == "" {
		c.Dir = defConf.Dir
	}

	if c.Port == "" {
		c.Port = defConf.Port
	}

	return &c
}

func (c *Conf) Default() {
	c.RootDir = "/var/graphqld"
	c.HotReload = false
	c.Dir = "/"
	c.Graphiql = false
	c.Port = "80"
}
