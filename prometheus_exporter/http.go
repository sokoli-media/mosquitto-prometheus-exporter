package prometheus_exporter

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"sync"
)

func RunHTTPServer(logger *slog.Logger, config MosquittoConfig) {
	var wg sync.WaitGroup
	quitMosquittoMetrics := make(chan bool, 1)

	wg.Add(1)
	go CollectMosquittoMetrics(logger, config, &wg, quitMosquittoMetrics)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/dashboard.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/dashboards/dashboard.json")
	})

	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		logger.Error("failed to run http server", "error", err)
		return
	}
}
