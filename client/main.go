package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

var (
	iterations = flag.Int("iterations", 0, "number of request to make to each endpoint")
	methods    = flag.String("methods", "", "which methods to call")
)

type Resonse struct {
	Message string `json:"message"`
}

func makeRequest(clientSpan opentracing.Span) {
	span := opentracing.StartSpan("make_request", opentracing.ChildOf(clientSpan.Context()))
	defer span.Finish()

	url := "http://server:8080/nocontext"
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	var response Resonse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		panic(err)
	}

	time.Sleep(time.Duration(getRandomInt(1, 5)) * time.Second)
	fmt.Printf("%+v\n", response)
}

func makeRequestWithContext(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "make_request_with_context")
	defer span.Finish()

	url := "http://server:8080/context"
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	var response Resonse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		panic(err)
	}

	time.Sleep(time.Duration(getRandomInt(1, 5)) * time.Second)
	fmt.Printf("%+v\n", response)
}

func getRandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

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

	clientSpan := tracer.StartSpan("client_span")
	defer clientSpan.Finish()

	var callMakeRequest bool
	var callMakeRequestWithContext bool

	for _, method := range strings.Split(*methods, ",") {
		if method == "makeRequest" {
			callMakeRequest = true
		}

		if method == "makeRequestWithContext" {
			callMakeRequestWithContext = true
		}
	}

	var wg sync.WaitGroup
	if callMakeRequest {
		for i := 0; i < *iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				makeRequest(clientSpan)
			}()
		}
	}

	if callMakeRequestWithContext {
		ctx := opentracing.ContextWithSpan(context.Background(), clientSpan)
		for i := 0; i < *iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				makeRequestWithContext(ctx)
			}()
		}
	}
	wg.Wait()
	fmt.Println("Done!")
}
