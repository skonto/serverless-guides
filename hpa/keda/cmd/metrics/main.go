package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	t       = false
	myGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge",
		Help: "Am I up?",
	})
)

func toggle(w http.ResponseWriter, r *http.Request) {
	if !t {
		myGauge.Set(5)
		t = true
	} else {
		myGauge.Set(0)
		t = false
	}
}

func ok(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "Hello!\n")
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/toggle", toggle)
	http.HandleFunc("/", ok)
	http.ListenAndServe(":8080", nil)
}
