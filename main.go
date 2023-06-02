package main

import (
	"expvar"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	go_kit_metrics "github.com/go-kit/kit/metrics"
	go_kit_expvar "github.com/go-kit/kit/metrics/expvar"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterMetrics(registry *GlobalRegistry, endpoint string) func(http.Handler) http.Handler {
	f := func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			registry.PromRegistry.HttpCounter.WithLabelValues(endpoint).Inc()

			expCounter := registry.ExpvarRegistry.GetCounter(endpoint)
			expCounter.Add(1)

			// Start measuring the duration
			start := time.Now()

			// Call the next handler
			next.ServeHTTP(w, r)

			registry.PromRegistry.HttpHistogram.WithLabelValues(endpoint).Observe(time.Since(start).Seconds())
		}
		return http.HandlerFunc(fn)
	}
	return f
}

type GlobalRegistry struct {
	PromRegistry   *PromRegistry
	ExpvarRegistry *ExpvarRegistry
}

func NewGlobalRegistry() *GlobalRegistry {
	return &GlobalRegistry{
		PromRegistry:   NewPromRegistry(),
		ExpvarRegistry: NewExpvarRegistry(),
	}
}

type ExpvarRegistry struct {
	metricsCounterVars map[string]go_kit_metrics.Counter
	counterLock        *sync.RWMutex
}

func NewExpvarRegistry() *ExpvarRegistry {
	return &ExpvarRegistry{
		metricsCounterVars: make(map[string]go_kit_metrics.Counter),
		counterLock:        &sync.RWMutex{},
	}
}

// GetCounter gets go-kit expvar Counter
func (m *ExpvarRegistry) GetCounter(key string) go_kit_metrics.Counter {

	if key == "" {
		fmt.Println("metrics counter key is empty")
		return nil
	}

	combinedKey := "counter" + "." + key

	m.counterLock.Lock()
	defer m.counterLock.Unlock()
	if val, ok := m.metricsCounterVars[combinedKey]; ok {
		return val
	}

	return m.createCounter(combinedKey)
}

func (m *ExpvarRegistry) createCounter(key string) *go_kit_expvar.Counter {
	counterVar := go_kit_expvar.NewCounter(key)
	m.metricsCounterVars[key] = counterVar
	return counterVar

}

type PromRegistry struct {
	HttpCounter   *prometheus.CounterVec
	HttpHistogram *prometheus.HistogramVec
}

func NewPromRegistry() *PromRegistry {
	// Create a new Prometheus counter metric
	registry := &PromRegistry{}
	registry.HttpCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint"},
	)

	registry.HttpHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "Duration of API requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// Register with Prometheus
	prometheus.MustRegister(registry.HttpCounter)
	prometheus.MustRegister(registry.HttpHistogram)
	return registry
}

func main() {
	// Create a new Chi router
	r := chi.NewRouter()

	counter := expvar.NewInt("counter")
	counter.Add(1)
	sample := expvar.NewString("name")
	sample.Set("value")

	// register prometheus metrics
	registry := NewGlobalRegistry()

	// Register middleware
	r.Use(middleware.Logger)

	// Register expvar handler
	r.Get("/debug/vars", expvarHandler)

	// Register prometheus metrics
	r.Get("/metrics", promMetricsHandler)

	// Two test handler
	r.With(RegisterMetrics(registry, "decide")).Get("/decide", decideHandler)
	r.With(RegisterMetrics(registry, "activate")).Get("/activate", activateHandler)

	// Add your other routes and handlers
	r.Get("/", helloHandler)

	// Start the server
	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", r)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func promMetricsHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	expvar.Handler().ServeHTTP(w, r)
}

func decideHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "decide handler")
}

func activateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "activate handler")
}
