package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type key uint

const (
	keyHeaderFunc key = iota
	keyHeader
	keyEnv
	keyCtxFile
	keyLog
)

func GetLogger(ctx context.Context) *zerolog.Logger {
	logger, ok := ctx.Value(keyLog).(*zerolog.Logger)
	if logger == nil || !ok {
		nop := zerolog.Nop()
		logger = &nop
	}

	return logger
}

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
	env, ok := ctx.Value(keyEnv).([]string)
	if !ok {
		return []string{}
	}

	return env
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
			c   = log.With().
				Str("mehod", r.Method).
				Str("host", r.Host).
				Str("url", r.URL.String()).
				Str("ip", r.RemoteAddr)
		)

		if ua := r.Header.Get("User-Agent"); ua != "" {
			c = c.Str("user-agent", ua)
		}

		if ref := r.Header.Get("Referer"); ref != "" {
			c = c.Str("referer", ref)
		}

		var logger = c.Logger()
		ctx = context.WithValue(ctx, keyLog, &logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromGraphConf(c config.GraphConf) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx    = r.Context()
				env    = getEnv(r)
				logger = GetLogger(ctx)
			)

			logger.Info().Send()

			ctx = context.WithValue(ctx, keyEnv, env)
			ctx = context.WithValue(ctx, keyHeaderFunc, w.Header)

			if cctx := c.Context; cctx != nil {
				ctxFile, err := ioutil.TempFile(cctx.TmpDir, "")
				if err != nil {
					logger.Error().Err(err).
						Msg("unable to create temporary context file")

					http.Error(w, err.Error(), http.StatusInternalServerError)

					return
				}
				defer func() {
					ctxFile.Close()
					os.Remove(ctxFile.Name())
				}()

				var ctxData []byte
				switch {
				case cctx.ExecPath != "":
					var (
						cmd = exec.Cmd{
							Path: cctx.ExecPath,
							Env:  env,
						}
					)

					if user := c.User; user != nil {
						cmd.SysProcAttr = &syscall.SysProcAttr{
							Credential: &syscall.Credential{
								Uid: user.Uid,
								Gid: user.Gid,
							},
						}
					}

					ctxData, err = cmd.Output()
					if err != nil {
						logger.Error().Err(err).
							Msg("unable to create a context from the ctx handler")

						http.Error(w, err.Error(), http.StatusInternalServerError)

						return
					}
				case cctx.Context != nil:
					var err error

					ctxData, err = json.Marshal(cctx.Context)
					if err != nil {
						logger.Error().Err(err).
							Msg("unable to marshal config context as JSON")

						http.Error(w, err.Error(), http.StatusInternalServerError)

						return
					}
				}

				if _, err := ctxFile.Write(ctxData); err != nil {
					logger.Error().Err(err).
						Msg("unable to write context to the context file")

					http.Error(w, err.Error(), http.StatusInternalServerError)

					return
				}

				ctx = context.WithValue(ctx, keyCtxFile, ctxFile)

			}

			r.Body = &limitedReaderCloser{
				LimitedReader: io.LimitedReader{
					R: r.Body,
					N: c.MaxBodyReadSize,
				},
			}

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getEnv(r *http.Request) []string {
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

type limitedReaderCloser struct {
	io.LimitedReader
}

func (lrc *limitedReaderCloser) Read(p []byte) (int, error) {
	return lrc.LimitedReader.Read(p)
}

func (lrc *limitedReaderCloser) Close() error {
	x, ok := lrc.R.(io.Closer)
	if !ok {
		return errors.New("unable to convert from type io.Reader to io.Closer in limitedReaderCloser.Close")
	}

	return x.Close()
}

func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				h      = r.Header.Get("Authorization")
				hParts = strings.Split(h, " ")
			)

			if len(hParts) < 2 {
				status := http.StatusUnauthorized
				http.Error(w, http.StatusText(status), status)
				return
			}

			if hParts[0] != "Basic" {
				status := http.StatusUnauthorized
				http.Error(w, http.StatusText(status), status)
				return
			}

			decodedBytes, err := base64.StdEncoding.DecodeString(hParts[1])
			if err != nil {
				status := http.StatusUnauthorized
				http.Error(w, http.StatusText(status), status)
				return
			}

			var authParts = strings.Split(string(decodedBytes), ":")
			if len(authParts) < 2 {
				status := http.StatusUnauthorized
				http.Error(w, http.StatusText(status), status)
				return
			}

			if authParts[0] != username || authParts[1] != password {
				status := http.StatusUnauthorized
				http.Error(w, http.StatusText(status), status)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
