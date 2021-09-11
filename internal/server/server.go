package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/raphaelreyna/graphqld/internal/config"
)

type Server struct {
	conf config.Conf

	graphs map[string]*server

	http.Server
}

func NewServer(conf config.Conf) (*Server, error) {
	var (
		mux = mux.NewRouter()

		s = Server{
			conf:   conf,
			graphs: make(map[string]*server),
		}
	)

	s.Addr = conf.Addr
	s.Handler = mux

	switch len(conf.Graphs) {
	case 1:
		gh, err := newServer(s.Addr, mux, s.conf.Graphs[0])
		if err != nil {
			return nil, err
		}

		s.graphs[""] = gh

	default:
		for _, gc := range s.conf.Graphs {
			mmux := mux.Host(gc.ServerName).Subrouter()

			gh, err := newServer(s.Addr, mmux, gc)
			if err != nil {
				return nil, err
			}

			s.graphs[gc.ServerName] = gh
		}
	}

	return &s, nil
}

func (s *Server) Start() error {
	for _, g := range s.graphs {
		if err := g.Start(); err != nil {
			return err
		}
	}

	switch tls := s.conf.TLS; tls {
	case nil:
		return s.Server.ListenAndServe()
	default:
		return s.Server.ListenAndServeTLS(tls.CertFile, tls.KeyFile)
	}

}

func (s *Server) Stop() error {
	for _, g := range s.graphs {
		g.Stop()
	}

	return s.Server.Close()
}
