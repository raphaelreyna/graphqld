package config

type BasicAuth struct {
	Username, Password string
}

func basicAuthFromMap(m map[interface{}]interface{}) *BasicAuth {
	return &BasicAuth{
		Username: m["username"].(string),
		Password: m["password"].(string),
	}
}
