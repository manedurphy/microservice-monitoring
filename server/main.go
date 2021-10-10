package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

var (
	httpDurationsSummary   *prometheus.SummaryVec
	httpDurationsHistogram *prometheus.HistogramVec
)

type Resonse struct {
	Message string `json:"message"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func(start time.Time) {
		httpDurationsSummary.With(prometheus.Labels{
			"func":     "handler",
			"endpoint": "http://server:8080/nocontext",
		}).Observe(float64(time.Since(start).Seconds()))

		httpDurationsHistogram.With(prometheus.Labels{
			"func":     "handler",
			"endpoint": "http://server:8080/nocontext",
		}).Observe(float64(time.Since(start).Seconds()))
	}(start)

	span, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

	serverSpan := opentracing.GlobalTracer().StartSpan("server_span", ext.RPCServerOption(span))
	defer serverSpan.Finish()
	time.Sleep(time.Duration(getRandomInt(1, 10)) * time.Second)

	resp := Resonse{
		Message: "request complete!",
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func handlerWithContext(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func(start time.Time) {
		httpDurationsSummary.With(prometheus.Labels{
			"func":     "handlerWithContext",
			"endpoint": "http://server:8080/context",
		}).Observe(float64(time.Since(start).Seconds()))

		httpDurationsHistogram.With(prometheus.Labels{
			"func":     "handlerWithContext",
			"endpoint": "http://server:8080/context",
		}).Observe(float64(time.Since(start).Seconds()))
	}(start)

	wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

	span := opentracing.StartSpan("server_span_with_context", ext.RPCServerOption(wireContext))
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	nestedCall(ctx)

	time.Sleep(time.Duration(getRandomInt(1, 10)) * time.Second)
	resp := Resonse{
		Message: "request w/context complete!",
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func nestedCall(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "nested_call_span")
	defer span.Finish()

	time.Sleep(time.Duration(getRandomInt(1, 10)) * time.Second)
}

func getRandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	httpDurationsSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "rpc_durations_summary_seconds",
		Help:       "RPC latency distributions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"func", "endpoint"})

	httpDurationsHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rpc_durations_histogram_seconds",
		Help:    "RPC latency distributions",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 12.5, 15.0, 17.5, 20.0},
	}, []string{"func", "endpoint"})

	prometheus.MustRegister(httpDurationsSummary)
	prometheus.MustRegister(httpDurationsHistogram)
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
}

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		panic(err)
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	defer closer.Close()
	if err != nil {
		panic(err)
	}
	opentracing.SetGlobalTracer(tracer)

	router := http.NewServeMux()
	router.HandleFunc("/nocontext", handler)
	router.HandleFunc("/context", handlerWithContext)
	router.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Starting server on port 8080...")
	log.Fatal(server.ListenAndServe())
}
