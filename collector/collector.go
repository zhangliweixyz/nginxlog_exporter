package collector

import (
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/satyrius/gonx"
	"github.com/zhangliweixyz/nginxlog_exporter/config"
	"log"
	"strings"
)

// Collector is a struct containing pointers to all metrics that should be
// exposed to Prometheus
type Collector struct {
	countTotal      *prometheus.CounterVec
	bytesTotal      *prometheus.CounterVec
	upstreamSeconds *prometheus.HistogramVec
	requestSeconds  *prometheus.HistogramVec

	staticValues    []string
	dynamicLabels   []string
	dynamicLabelLen int

	cfg    *config.AppConfig
	parser *gonx.Parser
}

// 创建Collector对象
// 每个应用配置都对应一个Collector
func NewCollector(cfg *config.AppConfig) *Collector {
	staticLables, staticValues := cfg.StaticLabelSets()
	dynamicLabels := cfg.DynamicLabels()

	labels := append(staticLables, dynamicLabels...)
	collector := &Collector{
		countTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Name,
			Name:      "http_request_count_total",
			Help:      "Amount of processed HTTP requests",
		}, labels),

		bytesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Name,
			Name:      "http_request_size_bytes",
			Help:      "Total amount of transferred bytes",
		}, labels),

		upstreamSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Name,
			Name:      "http_upstream_time_seconds",
			Help:      "Time needed by upstream servers to handle requests",
			Buckets:   cfg.HistogramBuckets,
		}, labels),

		requestSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Name,
			Name:      "http_request_time_seconds",
			Help:      "Time needed by NGINX to handle requests",
			Buckets:   cfg.HistogramBuckets,
		}, labels),

		staticValues:    staticValues,
		dynamicLabels:   dynamicLabels,
		dynamicLabelLen: len(dynamicLabels),

		cfg:    cfg,
		parser: gonx.NewParser(cfg.Format),
	}
	cfg.Prepare()

	return collector
}

// 采集解析原始数据，并更新相应指标值
func (c *Collector) Run() {
	// register to prometheus
	prometheus.MustRegister(c.countTotal)
	prometheus.MustRegister(c.bytesTotal)
	prometheus.MustRegister(c.upstreamSeconds)
	prometheus.MustRegister(c.requestSeconds)

	for _, f := range c.cfg.SourceFiles {
		t, err := tail.TailFile(f, tail.Config{
			Follow: true,
			ReOpen: true,
			Poll:   true,
		})
		if err != nil {
			log.Panic(err)
		}

		go func() {
			for line := range t.Lines {
				entry, err := c.parser.ParseString(line.Text)
				if err != nil {
					fmt.Printf("error while parsing line '%s': %s", line.Text, err)
					continue
				}

				dynamicValues := make([]string, c.dynamicLabelLen)
				for i, label := range c.dynamicLabels {
					if value, err := entry.Field(label); err == nil {
						dynamicValues[i] = c.formatValue(label, value)
					}
				}

				labelValues := append(c.staticValues, dynamicValues...)
				c.countTotal.WithLabelValues(labelValues...).Inc()

				if bytes, err := entry.FloatField("body_bytes_sent"); err == nil {
					c.bytesTotal.WithLabelValues(labelValues...).Add(bytes)
				}

				if times, err := entry.FloatField("upstream_response_time"); err == nil {
					c.upstreamSeconds.WithLabelValues(labelValues...).Observe(times)
				}

				if times, err := entry.FloatField("request_time"); err == nil {
					c.requestSeconds.WithLabelValues(labelValues...).Observe(times)
				}
			}
		}()
	}
}

func (c *Collector) formatValue(label, value string) string {
	replacement, ok := c.cfg.RelabelConfig.Replacement[label]
	if !ok {
		return value
	}

	if replacement.Trim != "" {
		value = strings.Split(value, replacement.Trim)[0]
	}

	for _, target := range replacement.Replaces {
		if target.Rex.MatchString(value) {
			return target.Value
		}
	}

	return value
}
