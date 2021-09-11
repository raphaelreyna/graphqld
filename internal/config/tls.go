package config

type TLS struct {
	CertFile, KeyFile string
}

func tlsFromMap(m map[string]interface{}) *TLS {
	return &TLS{
		CertFile: m["cert"].(string),
		KeyFile:  m["key"].(string),
	}
}
