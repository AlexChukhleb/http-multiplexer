package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) handleMultiplexer(w http.ResponseWriter, r *http.Request) {

	// ограничим кол-во запросов мультиплексора, на другие ручки ограничение не распространяется
	err := s.sem.Acquire(context.TODO(), time.Millisecond)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("number of requests exceeded"))
		return
	}
	defer s.sem.Release()

	ctx, cancelFunc := context.WithCancel(s.contextServer)
	defer cancelFunc()

	// надо добавить, но в тз не было
	//r.Body = http.MaxBytesReader(w, r.Body, 1024 * 1024)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var list []string
	{
		err := dec.Decode(&list)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("request parsing error"))
			return
		}
	}

	m, err := s.app.DoRequests(ctx, list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error creating response"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
