package cmd

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func CheckExists(config Config) error {
	reqObj, _ := http.NewRequest("GET", config.Endpoint, nil)
	if config.Authentication.Username != "" && config.Authentication.Password != "" {
		reqObj.SetBasicAuth(config.Authentication.Username, config.Authentication.Password)
	}
	resp, err := http.DefaultClient.Do(reqObj)
	if err != nil {
		log.Fatalf("Unable to send connection to PushGateway %s", err.Error())
		return err
	}

	if resp.StatusCode < 300 {
		log.Infof("Prometheus PushGateway Status %d", resp.StatusCode)
		return nil
	} else {
		log.Warnf("Prometheus PushGateway Status %d", resp.StatusCode)
		bytes, _ := io.ReadAll(resp.Body)
		log.Infof("%s", string(bytes))
		return err
	}
}

func SendMetrics(config Config, r io.ReadCloser, unique string) {
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		log.Errorf("Failed to read metrics: %v", err)
		return
	}

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(bytes.NewReader(data))
	if err != nil {
		log.Warnf("Error parsing metrics: %v", err)
		time.Sleep(10 * time.Second)
		return
	}

	pusher := push.New(config.Endpoint, config.PushGateway.Project).
		Grouping("instance", unique).
		BasicAuth(config.Authentication.Username, config.Authentication.Password)

	for name, mf := range metricFamilies {
		log.Debug("Processing metric: %s", name)
		if mf.Type == nil {
			log.Warnf("Metric %s has no type defined", name)
			continue
		}
		labelNames := getLabelNames(mf)
		switch *mf.Type {
		case dto.MetricType_GAUGE:
			pushMetric(pusher, createGaugeVec(name, mf, labelNames))
		case dto.MetricType_COUNTER:
			pushMetric(pusher, createCounterVec(name, mf, labelNames))
		case dto.MetricType_SUMMARY:
			pushMetric(pusher, createSummaryVec(name, mf, labelNames))
		case dto.MetricType_HISTOGRAM:
			pushMetric(pusher, createHistogramVec(name, mf, labelNames))
		default:
			log.Debug("Unsupported metric type: %v", *mf.Type)
		}
	}
	if err := pusher.Push(); err != nil {
		log.Errorf("Failed to push metrics: %v", err)
	} else {
		log.Infof("[%s] Metrics pushed to %s", unique, config.Endpoint)
	}
}

func getLabelNames(mf *dto.MetricFamily) []string {
	labelNames := map[string]struct{}{}

	for _, m := range mf.Metric {
		for _, label := range m.Label {
			labelNames[*label.Name] = struct{}{}
		}
	}
	var names []string
	for name := range labelNames {
		names = append(names, name)
	}
	return names
}

func pushMetric(pusher *push.Pusher, metric prometheus.Collector) {
	if metric != nil {
		pusher.Collector(metric)
	}
}

func createGaugeVec(name string, mf *dto.MetricFamily, labelNames []string) prometheus.Collector {
	metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: mf.GetHelp(),
		},
		labelNames,
	)
	for _, m := range mf.Metric {
		labels := extractLabels(m)
		standardizedLabels := standardizeLabels(labels, labelNames)
		if m.Gauge != nil {
			metric.With(standardizedLabels).Set(m.Gauge.GetValue())
		}
	}
	return metric
}

func createCounterVec(name string, mf *dto.MetricFamily, labelNames []string) prometheus.Collector {
	metric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: mf.GetHelp(),
		},
		labelNames,
	)
	for _, m := range mf.Metric {
		labels := extractLabels(m)
		standardizedLabels := standardizeLabels(labels, labelNames)
		if m.Counter != nil {
			metric.With(standardizedLabels).Add(m.Counter.GetValue())
		}
	}
	return metric
}

func createSummaryVec(name string, mf *dto.MetricFamily, labelNames []string) prometheus.Collector {
	metric := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: name,
			Help: mf.GetHelp(),
		},
		labelNames,
	)
	for _, m := range mf.Metric {
		labels := extractLabels(m)
		standardizedLabels := standardizeLabels(labels, labelNames)
		if m.Summary != nil {
			metric.With(standardizedLabels).Observe(m.Summary.GetSampleSum())
		}
	}
	return metric
}

func createHistogramVec(name string, mf *dto.MetricFamily, labelNames []string) prometheus.Collector {
	metric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: name,
			Help: mf.GetHelp(),
		},
		labelNames,
	)

	for _, m := range mf.Metric {
		labels := extractLabels(m)
		standardizedLabels := standardizeLabels(labels, labelNames)
		if m.Histogram != nil {

			for _, bucket := range m.Histogram.Bucket {
				metric.With(standardizedLabels).Observe(float64(bucket.GetUpperBound()))
			}
		}
	}
	return metric
}

func extractLabels(m *dto.Metric) prometheus.Labels {
	labels := prometheus.Labels{}
	for _, label := range m.Label {
		labels[*label.Name] = *label.Value
	}
	return labels
}

func standardizeLabels(labels prometheus.Labels, expected []string) prometheus.Labels {
	standardized := make(prometheus.Labels)
	for _, key := range expected {
		if val, ok := labels[key]; ok {
			standardized[key] = val
		} else {
			standardized[key] = ""
		}
	}
	return standardized
}
