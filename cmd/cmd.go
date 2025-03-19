package cmd

import (
	"flag"
	"time"

	"github.com/charmbracelet/log"
)

var (
	configPath            string
	PrometheusProxyConfig Config
)

func Run() error {
	flag.StringVar(&configPath, "config", "", "Set Configfile")
	flag.Parse()

	pushproxy, err := ReadConfig(configPath)
	if err != nil {
		log.Error(err.Error())
	}
	PrometheusProxyConfig = pushproxy
	if err := CheckExists(PrometheusProxyConfig); err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(PrometheusProxyConfig.NodeExporter.ScrapeTime) * time.Minute)
	defer ticker.Stop()
	log.Infof("Started Prometheus Push Gateway ticker (%d min interval)", PrometheusProxyConfig.NodeExporter.ScrapeTime)
	log.Info("Starting Initial Scrape!")
	go Scrape()
	for t := range ticker.C {
		log.Infof("Scraping! %s", t)
		Scrape()
	}
	return nil
}

func Scrape() {
	KubeAuth()
	FindNodeExportersInNamespace(PrometheusProxyConfig.Namespace)
	FindKubeStateMetricsInNamespace(PrometheusProxyConfig.Namespace)
	log.Info("-----------------")
}
