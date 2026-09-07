package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	dbOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_open_connections",
		Help: "Número de conexões abertas com o banco.",
	})
	dbInUseConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_in_use_connections",
		Help: "Número de conexões em uso no banco.",
	})
	dbIdleConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_idle_connections",
		Help: "Número de conexões ociosas no banco.",
	})
)

func UpdateDBMetrics(db *sql.DB) {
	stats := db.Stats()
	dbOpenConnections.Set(float64(stats.OpenConnections))
	dbInUseConnections.Set(float64(stats.InUse))
	dbIdleConnections.Set(float64(stats.Idle))
}
