package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
)

var Config Conf

type GraphConf struct {
	ServerName      string
	DocumentRoot    string
	HotReload       bool
	hotReloadSet    bool
	ResolverDir     string
	Graphiql        bool
	graphiqlSet     bool
	ContextExecPath string
	ContextFilesDir string
}

type Conf struct {
	Hostname        string
	Addr            string
	RootDir         string
	HotReload       bool
	ResolverDir     string
	Graphiql        bool
	ContextExecPath string
	ContextFilesDir string
	Graphs          []GraphConf
}

func init() {
	viper.SetConfigName("graphqld")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/graphqld/")
	viper.AddConfigPath("/etc/")

	viper.SetEnvKeyReplacer(strings.NewReplacer(
		"CONTEXTEXECPATH", "CTX_EXEC_PATH",
		"CONTEXTFILESDIR", "CTX_FILES_DIR",
		"RESOLVERDIR", "RESOLVER_DIR",
		"LOGJSON", "LOG_JSON",
		"LOGCOLOR", "LOG_COLOR",
	))

	viper.SetEnvPrefix("GRAPHQLD")
	viper.AutomaticEnv()

	viper.SetDefault("hostname", "")
	viper.SetDefault("address", "")
	viper.SetDefault("root", "/var/graphqld")
	viper.SetDefault("hot", false)
	viper.SetDefault("graphiql", false)
	viper.SetDefault("contextExecPath", "")
	viper.SetDefault("contextFilesDir", "")
	viper.SetDefault("resolverDir", "/")
	viper.SetDefault("port", "80")
	viper.SetDefault("logJSON", false)
	viper.SetDefault("logColor", true)
	viper.SetDefault("resolverDir", "/")

	if !viper.GetBool("logJSON") {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: !viper.GetBool("logColor"),
		})
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal().Err(err).
				Str("configuration-file", viper.ConfigFileUsed()).
				Msg("error reading configuration file")
		}

		log.Info().Msg("no configuration file found")
	} else {
		log.Info().
			Str("file", viper.ConfigFileUsed()).
			Msg("read configuration")
	}

	Config.Hostname = viper.GetString("hostname")
	Config.Addr = viper.GetString("address")
	Config.RootDir = viper.GetString("root")
	Config.HotReload = viper.GetBool("hot")
	Config.Graphiql = viper.GetBool("graphiql")
	Config.ResolverDir = viper.GetString("resolverDir")

	if !filepath.IsAbs(Config.RootDir) {
		path, err := filepath.Abs(Config.RootDir)
		if err != nil {
			log.Fatal().Err(err).
				Str("path", Config.RootDir).
				Msg("unable to compute absolute rooth path")
		}
		Config.RootDir = path
	}
	if !filepath.IsAbs(Config.ResolverDir) {
		path, err := filepath.Abs(Config.ResolverDir)
		if err != nil {
			log.Fatal().Err(err).
				Str("path", Config.ResolverDir).
				Msg("unable to compute resolver dir root path")
		}
		Config.ResolverDir = path
	}

	if Config.Addr == "" {
		Config.Addr = ":" + viper.GetString("port")
	}

	var (
		confGraphs = make(map[string]GraphConf)
		dirGraphs  = make([]GraphConf, 0)
	)
	{
		if iface := viper.Get("graphs"); iface != nil {
			for _, v := range iface.([]interface{}) {
				var (
					m  = v.(map[interface{}]interface{})
					gc GraphConf
				)

				if x, ok := m["serverName"]; ok {
					gc.ServerName = x.(string)
				}

				if x, ok := m["hot"]; ok {
					gc.HotReload = x.(bool)
					gc.hotReloadSet = true
				}

				if x, ok := m["resolverDir"]; ok {
					gc.ResolverDir = x.(string)
				}

				if x, ok := m["graphiql"]; ok {
					gc.Graphiql = x.(bool)
					gc.graphiqlSet = true
				}

				if x, ok := m["contextExecPath"]; ok {
					gc.ContextExecPath = x.(string)
				}

				if x, ok := m["contextFilesDir"]; ok {
					gc.ContextFilesDir = x.(string)
				}

				if !filepath.IsAbs(gc.ResolverDir) && gc.ResolverDir != "" {
					path, err := filepath.Abs(gc.ResolverDir)
					if err != nil {
						log.Fatal().Err(err).
							Str("path", gc.ResolverDir).
							Msg("unable to compute resolver dir root path")
					}
					gc.ResolverDir = path
				}
				if !filepath.IsAbs(gc.ContextExecPath) && gc.ContextExecPath != "" {
					path, err := filepath.Abs(gc.ContextExecPath)
					if err != nil {
						log.Fatal().Err(err).
							Str("path", gc.ContextExecPath).
							Msg("unable to compute absolute rooth path")
					}
					gc.ContextExecPath = path
				}
				if !filepath.IsAbs(gc.ContextFilesDir) && gc.ContextFilesDir != "" {
					path, err := filepath.Abs(gc.ContextFilesDir)
					if err != nil {
						log.Fatal().Err(err).
							Str("path", gc.ContextFilesDir).
							Msg("unable to compute absolute rooth path")
					}
					gc.ContextFilesDir = path
				}

				confGraphs[gc.ServerName] = gc
			}
		}

		dirs, err := os.ReadDir(Config.RootDir)
		if err != nil {
			log.Fatal().Err(err).
				Msg("unable to read root directory")
		}

		for _, item := range dirs {
			if !item.IsDir() {
				continue
			}

			var (
				name = item.Name()
				path = filepath.Join(Config.RootDir, name)
			)

			isOk, err := isGraphDir(path)
			if err != nil {
				log.Fatal().Err(err).
					Msg("unable to check if directory has graph")
			}

			if !isOk {
				continue
			}

			gc := GraphConf{
				HotReload:       viper.GetBool("hot"),
				Graphiql:        viper.GetBool("graphiql"),
				DocumentRoot:    path,
				ResolverDir:     viper.GetString("resolverDir"),
				ContextExecPath: viper.GetString("contextExecPath"),
				ContextFilesDir: viper.GetString("contextFilesDir"),
			}

			if err := checkDomain(name); err != nil {
				log.Fatal().Err(err).
					Msg("invalid domain name")
			}
			gc.ServerName = name

			dirGraphs = append(dirGraphs, gc)
		}

		if len(dirGraphs) == 0 {
			gc := GraphConf{
				HotReload:       viper.GetBool("hot"),
				Graphiql:        viper.GetBool("graphiql"),
				DocumentRoot:    Config.RootDir,
				ResolverDir:     viper.GetString("resolverWD"),
				ContextExecPath: viper.GetString("contextExecPath"),
				ContextFilesDir: viper.GetString("contextFilesDir"),
			}

			if Config.Hostname != "" {
				if err := checkDomain(Config.Hostname); err != nil {
					log.Fatal().Err(err).
						Msg("invalid domain name")
				}
			}

			dirGraphs = append(dirGraphs, gc)
		}
	}

	Config.Graphs = make([]GraphConf, 0)
	for _, graph := range dirGraphs {
		confGraph, ok := confGraphs[graph.ServerName]
		if !ok {
			Config.Graphs = append(Config.Graphs, graph)
			continue
		}

		if x := confGraph.ResolverDir; x != "" {
			graph.ResolverDir = x
		}

		if x := confGraph.HotReload; confGraph.hotReloadSet {
			graph.HotReload = x
		}

		if x := confGraph.Graphiql; confGraph.graphiqlSet {
			graph.Graphiql = x
		}

		if x := confGraph.ContextExecPath; x != "" {
			graph.ContextExecPath = x
		}

		if x := confGraph.ContextFilesDir; x != "" {
			graph.ContextFilesDir = x
		}

		Config.Graphs = append(Config.Graphs, graph)
	}
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

func isGraphDir(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		var name = entry.Name()

		if name == "Query" || name == "Mutation" {
			return true, nil
		}
	}

	return false, nil
}
