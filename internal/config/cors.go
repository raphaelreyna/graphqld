package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type CORSConfig struct {
	AllowCredentials bool
	AllowedHeaders   []string
	AllowedOrigins   []string
	IgnoreOptions    bool
}

func CORSConfigFromViper() *CORSConfig {
	var cc CORSConfig
	if stringMap := viper.GetStringMap("cors"); stringMap != nil {
		if x, ok := stringMap["allowCredentials"]; ok {
			x, ok := x.(bool)
			if !ok {
				log.Fatal().
					Msg("cors.allowCredentials expected bool")
			}
			cc.AllowCredentials = x
		}

		if x, ok := stringMap["ignoreOptions"]; ok {
			x, ok := x.(bool)
			if !ok {
				log.Fatal().
					Msg("cors.ignoreOptions expected bool")
			}
			cc.IgnoreOptions = x
		}

		if x, ok := stringMap["allowedHeaders"]; ok {
			ifaces, ok := x.([]interface{})
			if !ok {
				log.Fatal().
					Msg("cors.allowedHeaders expected []string")
			}

			var headers = make([]string, 0)
			for _, iface := range ifaces {
				header, ok := iface.(string)
				if !ok {
					log.Fatal().
						Msg("cors.allowedHeaders expected []string")
				}
				headers = append(headers, header)
			}
			cc.AllowedHeaders = headers
		}

		if x, ok := stringMap["allowedOrigins"]; ok {
			ifaces, ok := x.([]interface{})
			if !ok {
				log.Fatal().
					Msg("cors.allowedOrigins expected []string")
			}

			var origins = make([]string, 0)
			for _, iface := range ifaces {
				origin, ok := iface.(string)
				if !ok {
					log.Fatal().
						Msg("cors.allowedOrigins expected []string")
				}
				origins = append(origins, origin)
			}
			cc.AllowedOrigins = origins
		}

		return &cc
	}

	return nil
}

func CORSConfigFromMap(m map[interface{}]interface{}) *CORSConfig {
	var cc CORSConfig

	if x, ok := m["allowCredentials"].(bool); ok {
		cc.AllowCredentials = x
	}

	if x, ok := m["ignoreOptions"].(bool); ok {
		cc.IgnoreOptions = x
	}

	if x, ok := m["allowedHeaders"].([]interface{}); ok {
		var allowedHeaders = make([]string, len(x))
		for idx, header := range x {
			hstr, ok := header.(string)
			if !ok {
				continue
			}
			allowedHeaders[idx] = hstr
		}
		cc.AllowedHeaders = allowedHeaders
	}

	if x, ok := m["allowedOrigins"].([]interface{}); ok {
		var allowedOrigins = make([]string, len(x))
		for idx, origin := range x {
			hstr, ok := origin.(string)
			if !ok {
				continue
			}
			allowedOrigins[idx] = hstr
		}
		cc.AllowedOrigins = allowedOrigins
	}

	return &cc
}
