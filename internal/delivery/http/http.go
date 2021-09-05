package http

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ach/http-multiplexer/internal/app"
	"github.com/ach/http-multiplexer/pkg/semaphore"
)

type Server struct {
	config        *Config
	app           *app.App
	sem           semaphore.Semaphore
	contextServer context.Context
}

func NewServer(config *Config, app *app.App) *Server {
	return &Server{
		config: config,
		app:    app,
		sem:    semaphore.NewSemaphore(config.MaxIncomingConnections),
	}
}

func (s *Server) Run(ctx context.Context) error {

	s.contextServer = ctx

	router := http.NewServeMux()
	router.HandleFunc("/", s.handleMultiplexer)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(s.config.Port),
		Handler: router,
	}

	var err error
	go func() {
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Printf("server started")

	<-ctx.Done()

	log.Printf("server stopped")

	ctxShutDown, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer func() {
		cancelFunc()
	}()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown Failed:%+s", err)
	}

	log.Printf("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return err
}
