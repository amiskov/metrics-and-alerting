package api

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

var indexTmpl = template.Must(
	template.New("index").Parse(`<h1>Metrics</h1>
		<table>
		{{range $m := .Metrics}}
			 <tr>
			 <td>{{$m.ID}}</td>
			 	<td>{{$m.MType}}</td>

			 {{if (eq $m.MType "gauge")}}
			 	<td>{{$m.Value}}</td>
			 {{ end }}

			 {{if (eq $m.MType "counter")}}
			 	<td>{{$m.Delta}}</td>
			 {{ end }}
			 	<td>{{$m.Hash}}</td>
			 </tr>
		{{end}}
		</table>`))

func (api *metricsAPI) getMetricsList(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	metrics, err := api.repo.GetAll()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		logger.Log(r.Context()).Errorf("failed getting metrics: %v", err)
		return
	}

	err = indexTmpl.Execute(rw,
		struct {
			Metrics []models.Metrics
		}{
			Metrics: metrics,
		})

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		logger.Log(r.Context()).Errorf("failed executing template: %v", err)
		return
	}
}

func (api *metricsAPI) getMetric(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	m, err := api.repo.Get(metricType, metricName)
	if err != nil {
		logger.Log(r.Context()).Errorf("Metric not found. Body: %s. Error: %v.", err)
		rw.WriteHeader(http.StatusNotFound)
		writeBody(r.Context(), rw, []byte(`Can't get metric `+err.Error()))
		return
	}

	strVal, err := m.GetStrVal()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(r.Context(), rw, []byte(err.Error()))
		return
	}

	rw.WriteHeader(http.StatusOK)
	writeBody(r.Context(), rw, []byte(strVal))
}

func (api *metricsAPI) upsertMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")

	urlVal := chi.URLParam(r, "metricValue")
	mType := chi.URLParam(r, "metricType")

	metricData := models.Metrics{
		MType: mType,
		ID:    chi.URLParam(r, "metricName"),
	}

	switch mType {
	case models.MCounter:
		delta, err := strconv.ParseInt(urlVal, 10, 64)
		if err != nil {
			logger.Log(r.Context()).Errorf("failed parsing counter delta: %v", err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metricData.Delta = &delta
	case models.MGauge:
		val, err := strconv.ParseFloat(urlVal, 64)
		if err != nil {
			logger.Log(r.Context()).Errorf("failed parsing gauge value: %v", err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metricData.Value = &val
	}

	err := api.repo.Update(metricData)
	switch {
	case errors.Is(err, models.ErrorBadMetricFormat):
		rw.WriteHeader(http.StatusBadRequest)
		return
	case errors.Is(err, models.ErrorMetricNotFound):
		rw.WriteHeader(http.StatusNotFound)
		return
	case errors.Is(err, models.ErrorUnknownMetricType):
		rw.WriteHeader(http.StatusNotImplemented)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func handleNotFound(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusNotFound)
	writeBody(r.Context(), rw, []byte("not found"))
}

func (api *metricsAPI) ping(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")

	err := api.repo.Ping(r.Context())
	if err != nil {
		logger.Log(r.Context()).Errorf("can't connect to DB: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		writeBody(r.Context(), rw, []byte("DB connection failed"))
		return
	}

	rw.WriteHeader(http.StatusOK)
	writeBody(r.Context(), rw, []byte("DB connected successfully"))
}

func handleNotImplemented(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusNotImplemented)
	writeBody(r.Context(), rw, []byte("not implemented"))
}

func writeBody(ctx context.Context, rw http.ResponseWriter, body []byte) {
	_, err := rw.Write(body)
	if err != nil {
		logger.Log(ctx).Errorf("Failed writing response body: %v", err)
	}
}
