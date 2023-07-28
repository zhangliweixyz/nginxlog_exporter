package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhangliweixyz/nginxlog_exporter/collector"
	"github.com/zhangliweixyz/nginxlog_exporter/config"
	"log"
	"net/http"
)

var (
	bind, configFile string
)

func main() {
	flag.StringVar(&bind, "web.listen-address", ":9999", "Address to listen on for the web interface and API.")
	flag.StringVar(&configFile, "config.file", "config.yml", "Nginx log exporter configuration file name.")
	flag.Parse()

	cfg, err := config.LoadFile(configFile)
	if err != nil {
		log.Panic(err)
	}

	for _, app := range cfg.Apps {
		go collector.NewCollector(app).Run()
	}

	fmt.Printf("running HTTP server on address %s\n", bind)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Nginx Log Exporter</title></head>
             <body>
             <h1>Nginx Log Exporter</h1>
             <p><a href=/metrics>metrics</a></p>
             </body>
             </html>`))
	})

	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Fatalf("start server with error: %v\n", err)
	}
}
