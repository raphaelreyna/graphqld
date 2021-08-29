package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

type key uint

const (
	keyHeaderFunc key = iota
	keyHeader
	keyEnv
	keyCtxFile
)

func GetCtxFile(ctx context.Context) *os.File {
	file, _ := ctx.Value(keyCtxFile).(*os.File)
	return file
}

func GetWHeader(ctx context.Context) http.Header {
	headerFunc := ctx.Value(keyHeaderFunc).(func() http.Header)
	return headerFunc()
}

func GetRHeader(ctx context.Context) http.Header {
	return ctx.Value(keyHeader).(http.Header)
}

func GetEnv(ctx context.Context) []string {
	return ctx.Value(keyEnv).([]string)
}

func getEnv(port string, r *http.Request) []string {
	var (
		upperCaseAndUnderscore = func(r rune) rune {
			switch {
			case r >= 'a' && r <= 'z':
				return r - ('a' - 'A')
			case r == '-':
				return '_'
			case r == '=':
				return '_'
			}
			return r
		}

		removeLeadingDuplicates = func(env []string) (ret []string) {
			for i, e := range env {
				found := false
				if eq := strings.IndexByte(e, '='); eq != -1 {
					keq := e[:eq+1]
					for _, e2 := range env[i+1:] {
						if strings.HasPrefix(e2, keq) {
							found = true
							break
						}
					}
				}
				if !found {
					ret = append(ret, e)
				}
			}
			return
		}
	)

	env := []string{
		"SERVER_SOFTWARE=graphqld",
		"SERVER_NAME=" + r.Host,
		"SERVER_PROTOCOL=HTTP/1.1",
		"HTTP_HOST=" + r.Host,
		"GATEWAY_INTERFACE=CGGI/1.1",
		"REQUEST_URI=" + r.URL.RequestURI(),
		"SERVER_PORT=" + port,
	}
	if remoteIP, remotePort, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		env = append(env, "REMOTE_ADDR="+remoteIP, "REMOTE_HOST="+remoteIP, "REMOTE_PORT="+remotePort)
	} else {
		env = append(env, "REMOTE_ADDR="+r.RemoteAddr, "REMOTE_HOST="+r.RemoteAddr)
	}
	if r.TLS != nil {
		env = append(env, "HTTPS=on")
	}

	for k, v := range r.Header {
		k = strings.Map(upperCaseAndUnderscore, k)
		if k == "PROXY" {
			continue
		}
		joinStr := ", "
		if k == "COOKIE" {
			joinStr = "; "
		}
		env = append(env, "HTTP_"+k+"="+strings.Join(v, joinStr))
	}

	if r.ContentLength > 0 {
		env = append(env, fmt.Sprintf("CONTENT_LENGTH=%d", r.ContentLength))
	}
	if ctype := r.Header.Get("Content-Type"); ctype != "" {
		env = append(env, "CONTENT_TYPE="+ctype)
	}

	envPath := os.Getenv("PATH")
	if envPath == "" {
		envPath = "/bin:/usr/bin:/usr/ucb:/usr/bsd:/usr/local/bin"
	}
	env = append(env, "PATH="+envPath)

	return removeLeadingDuplicates(env)
}
