package config

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"gopkg.in/yaml.v2"
)

type GraphConf struct {
	ServerName      string `yaml:"ServerName"`
	DocumentRoot    string `yaml:"DocumentRoot"`
	HotReload       bool   `yaml:"HotReload"`
	ResolverDir     string `yaml:"ResolverDir"`
	Graphiql        bool   `yaml:"GraphiQL"`
	ContextExecPath string `yaml:"ContextExec"`
	ContextFilesDir string `yaml:"ContextFilesDir"`
}

type Conf struct {
	Hostname string      `yaml:"Hostname"`
	Addr     string      `yaml:"Address"`
	RootDir  string      `yaml:"RootDir"`
	Graphs   []GraphConf `yaml:"Graphs"`
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

func ParseFromEnv() (*Conf, error) {
	var (
		c = Conf{
			Hostname: os.Getenv("GRAPHQLD_HOSTNAME"),
			RootDir:  os.Getenv("GRAPHQLD_ROOT_DIR"),
			Addr:     os.Getenv("GRAPHQLD_ADDRESS"),
			Graphs:   make([]GraphConf, 0),
		}

		defConf = Conf{}

		hotReload   bool
		graphiql    bool
		dir         = os.Getenv("GRAPHQLD_RESOLVER_DIR")
		ctxExecPath = os.Getenv("GRAPHQLD_CTX_EXEC")
		ctxFilesDir = os.Getenv("GRAPHQLD_CTX_DIR")
		port        = os.Getenv("GRAPHQLD_PORT")
	)

	defConf.Default()

	switch x := os.Getenv("GRAPHQLD_HOT_RELOAD"); x {
	case "":
		hotReload = false
	default:
		hotReload = true
	}

	switch x := os.Getenv("GRAPHQLD_GRAPHIQL"); x {
	case "":
		graphiql = false
	default:
		graphiql = true
	}

	if dir == "" {
		dir = "/"
	}

	if port != "" {
		c.Addr = ":" + port
	}

	if c.RootDir == "" {
		c.RootDir = defConf.RootDir
	}

	dirs, err := os.ReadDir(c.RootDir)
	if err != nil {
		return nil, nil
	}

	for _, item := range dirs {
		if !item.IsDir() {
			c.Graphs = make([]GraphConf, 0)
			break
		}

		name := item.Name()

		gc := GraphConf{
			HotReload:       hotReload,
			Graphiql:        graphiql,
			DocumentRoot:    filepath.Join(c.RootDir, name),
			ResolverDir:     dir,
			ContextExecPath: ctxExecPath,
			ContextFilesDir: ctxFilesDir,
		}

		if err := checkDomain(name); err != nil {
			return nil, err
		}
		gc.ServerName = name

		c.Graphs = append(c.Graphs, gc)
	}

	if len(c.Graphs) == 0 {
		gc := GraphConf{
			HotReload:       hotReload,
			Graphiql:        graphiql,
			DocumentRoot:    c.RootDir,
			ResolverDir:     dir,
			ContextExecPath: ctxExecPath,
			ContextFilesDir: ctxFilesDir,
		}

		if c.Hostname != "" {
			if err := checkDomain(c.Hostname); err != nil {
				return nil, err
			}
		}

		c.Graphs = append(c.Graphs, gc)
	}

	return &c, nil
}

func (c *Conf) Default() {
	c.RootDir = "/var/graphqld"
	c.Addr = ":80"
}

// Copy-Pasted from: https://gist.github.com/chmike/d4126a3247a6d9a70922fc0e8b4f4013
func checkDomain(name string) error {
	switch {
	case len(name) == 0:
		return nil // an empty domain name will result in a cookie without a domain restriction
	case len(name) > 255:
		return fmt.Errorf("invalid domain: name length is %d, can't exceed 255", len(name))
	}
	var l int
	for i := 0; i < len(name); i++ {
		b := name[i]
		if b == '.' {
			// check domain labels validity
			switch {
			case i == l:
				return fmt.Errorf("invalid domain: invalid character '%c' at offset %d: label can't begin with a period", b, i)
			case i-l > 63:
				return fmt.Errorf("invalid domain: byte length of label '%s' is %d, can't exceed 63", name[l:i], i-l)
			case name[l] == '-':
				return fmt.Errorf("invalid domain: label '%s' at offset %d begins with a hyphen", name[l:i], l)
			case name[i-1] == '-':
				return fmt.Errorf("invalid domain: label '%s' at offset %d ends with a hyphen", name[l:i], l)
			}
			l = i + 1
			continue
		}
		// test label character validity, note: tests are ordered by decreasing validity frequency
		if !(b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '-' || b >= 'A' && b <= 'Z') {
			// show the printable unicode character starting at byte offset i
			c, _ := utf8.DecodeRuneInString(name[i:])
			if c == utf8.RuneError {
				return fmt.Errorf("invalid domain: invalid rune at offset %d", i)
			}
			return fmt.Errorf("invalid domain: invalid character '%c' at offset %d", c, i)
		}
	}
	// check top level domain validity
	switch {
	case l == len(name):
		return fmt.Errorf("invalid domain: missing top level domain, domain can't end with a period")
	case len(name)-l > 63:
		return fmt.Errorf("invalid domain: byte length of top level domain '%s' is %d, can't exceed 63", name[l:], len(name)-l)
	case name[l] == '-':
		return fmt.Errorf("invalid domain: top level domain '%s' at offset %d begins with a hyphen", name[l:], l)
	case name[len(name)-1] == '-':
		return fmt.Errorf("invalid domain: top level domain '%s' at offset %d ends with a hyphen", name[l:], l)
	case name[l] >= '0' && name[l] <= '9':
		return fmt.Errorf("invalid domain: top level domain '%s' at offset %d begins with a digit", name[l:], l)
	}
	return nil
}
