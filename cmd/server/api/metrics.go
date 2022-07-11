package api

import (
	"html/template"
	"log"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

func (api *api) GetIndex(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("index").Parse(`<h1>Metrics Service</h1>
		<h2>Gauge Metrics</h2>
		<table>
		{{range $m := .GaugeMetrics}}
			 <tr><td>{{$m.Name}}</td><td>{{$m.Value}}</td></tr>
		{{end}}
		</table>
		<h2>Counter Metrics</h2>
		<table>
		{{range $m := .CounterMetrics}}
			 <tr><td>{{$m.Name}}</td><td>{{$m.Value}}</td></tr>
		{{end}}
		</table>`)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(rw,
		struct {
			GaugeMetrics   []models.MetricRaw
			CounterMetrics []models.MetricRaw
		}{
			api.store.GetGaugeMetrics(),
			api.store.GetCounterMetrics(),
		})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("error while executing the template")
	}
}
