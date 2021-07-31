package http

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HTTPServer struct {
	addr     string
	server   *http.Server
	dieChan  chan struct{}
	running  bool
	errGroup *errgroup.Group
}

func (h *HTTPServer) Start() {

	defer func() {
		h.running = false
	}()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	h.errGroup.Go(func() error {
		var err error
		select {
		case <-h.dieChan:
			err = errors.New("the server will shutdown in a few seconds")
		case s := <-sg:
			close(h.dieChan)
			err = errors.New(fmt.Sprintf("got signal: %s shutting down...", s.String()))
		}
		err2 := h.server.Close()
		if err2 != nil && err2 != http.ErrServerClosed {
			return fmt.Errorf("server close error:%w", err2)
		}
		return err
	})

	h.errGroup.Go(func() error {
		err := h.listenAndServe()
		return err
	})

	err := h.errGroup.Wait()
	if err != nil {
		println("errors:", err.Error())
	}
}

func (h *HTTPServer) ShutDown() {
	select {
	case <-h.dieChan:
	default:
		close(h.dieChan)
	}
}

func (h *HTTPServer) GetDieChan() chan struct{} {
	return h.dieChan
}

func (h *HTTPServer) listenAndServe() error {
	err := h.server.ListenAndServe()
	if err != nil {
		return errors.New(fmt.Sprintf("listen and serve err:%v", err.Error()))
	}
	h.running = true
	return nil
}

func NewHttpServer() *HTTPServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello world", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})
	addr := ":8080"
	s := &http.Server{
		Addr:         addr,
		Handler:      mux,
		WriteTimeout: 3 * time.Second,
	}
	return &HTTPServer{addr: addr, server: s, dieChan: make(chan struct{}), errGroup: new(errgroup.Group)}
}
