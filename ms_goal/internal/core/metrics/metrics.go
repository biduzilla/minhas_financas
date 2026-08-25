package metrics

import (
	"database/sql"
	"net/http"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requisições HTTP processadas, particionado por método, path e status.",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duração das requisições HTTP em segundos.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	goGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_goroutines_custom",
		Help: "Número de goroutines ativas.",
	})

	goThreads = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_threads_custom",
		Help: "Número de threads do sistema operacional criados.",
	})

	goMemoryAlloc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_memory_alloc_bytes",
		Help: "Número de bytes de memória alocados e ainda em uso.",
	})

	goMemoryTotalAlloc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_memory_total_alloc_bytes",
		Help: "Número total de bytes alocados (incluindo os já liberados).",
	})

	goMemorySys = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_memory_sys_bytes",
		Help: "Número total de bytes de memória obtidos do sistema operacional.",
	})

	goGCCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_gc_count",
		Help: "Número de ciclos de garbage collection completados.",
	})
)

var (
	LoginAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_attempts_total",
			Help: "Total de tentativas de login, por resultado.",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(LoginAttempts)

	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)

	prometheus.MustRegister(dbOpenConnections)
	prometheus.MustRegister(dbInUseConnections)
	prometheus.MustRegister(dbIdleConnections)

	prometheus.MustRegister(goGoroutines)
	prometheus.MustRegister(goThreads)
	prometheus.MustRegister(goMemoryAlloc)
	prometheus.MustRegister(goMemoryTotalAlloc)
	prometheus.MustRegister(goMemorySys)
	prometheus.MustRegister(goGCCount)
}

func UpdateGoroutinesMetrics() {
	goGoroutines.Set(float64(runtime.NumGoroutine()))
	goThreads.Set(float64(runtime.NumCPU()))

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	goMemoryAlloc.Set(float64(memStats.Alloc))
	goMemoryTotalAlloc.Set(float64(memStats.TotalAlloc))
	goMemorySys.Set(float64(memStats.Sys))
	goGCCount.Set(float64(memStats.NumGC))
}

func MetricsHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UpdateDBMetrics(db)
		UpdateGoroutinesMetrics()
		promhttp.Handler().ServeHTTP(w, r)
	})
}
