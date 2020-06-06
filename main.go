package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	ListenPort string
}

func initConfig() Config {
	v := viper.New()
	v.AutomaticEnv()

	_ = v.BindEnv("listenport", "LISTEN_PORT")
	v.SetDefault("listenport", ":8080")

	var conf Config
	if err := v.Unmarshal(&conf); err != nil {
		panic(err)
	}

	if !strings.HasPrefix(conf.ListenPort, ":") {
		conf.ListenPort = ":" + conf.ListenPort
	}

	return conf
}

func helloHandler(summary *prometheus.SummaryVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()
		defer func() {
			duration := time.Since(start)
			summary.WithLabelValues("duration").Observe(duration.Seconds())
		}()

		name := chi.URLParam(r, "name")

		if name == "simulate" {
			time.Sleep(10 * time.Second)
		}

		_, err := w.Write([]byte(fmt.Sprintf("Hello %s", name)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func NewRouter(summary *prometheus.SummaryVec) *chi.Mux {
	router := chi.NewMux()
	router.Get("/hello/{name}", helloHandler(summary))
	router.Get("/healthz", healthHandler)
	router.Mount("/metrics", promhttp.Handler())
	router.HandleFunc("/*", defaultHandler)
	return router
}

func main() {
	conf := initConfig()

	helloHandlerSummary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "hello_latency",
		Help: "Time taken for greeting request",
	}, []string{"hello"})

	httpHandler := NewRouter(helloHandlerSummary)

	httpServer := &http.Server{
		Handler: httpHandler,
	}

	err := prometheus.Register(helloHandlerSummary)
	if err != nil {
		log.Fatal("Error registering Prometheus summary")
	}

	httpListener, err := net.Listen("tcp", conf.ListenPort)
	if err != nil {
		log.Fatal("Error listening on tcp port 8080")
	}

	var group run.Group
	group.Add(
		func() error {
			log.Infof("starting server on port %s", conf.ListenPort)

			return httpServer.Serve(httpListener)
		},
		func(e error) {
			log.Info("shutting server down")

			ctx := context.Background()
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(5*time.Second))
			defer cancel()

			err := httpServer.Shutdown(ctx)
			if err != nil {
				log.Error(err)
			}

			_ = httpServer.Close()
		},
	)

	{
		var (
			cancelInterrupt = make(chan struct{})
			ch              = make(chan os.Signal, 2)
		)
		defer close(ch)

		group.Add(
			func() error {
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
				select {
				case <-ch:
				case <-cancelInterrupt:
				}

				return nil
			},
			func(e error) {
				close(cancelInterrupt)
				signal.Stop(ch)
			},
		)
	}

	if err := group.Run(); err != nil {
		log.Fatal("run group failed", zap.Error(err))
	}
}
