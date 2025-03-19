package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset

func KubeAuth() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func FetchMetricsFromNodeExporter(url string, namespace string) (io.ReadCloser, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s.%s.pod.cluster.local:9100/metrics", strings.ReplaceAll(url, ".", "-"), namespace))
	if err != nil {
		log.Warnf("Unable to fetch node exporter metrics : %s", err.Error())
		return io.NopCloser(bytes.NewBuffer([]byte{})), err
	}
	return resp.Body, nil
}

func FindNodeExportersInNamespace(namespace string) {
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Error(err.Error())
	}

	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, "node-exporter") {
			log.Info(pod.Name)
			metrics, err := FetchMetricsFromNodeExporter(pod.Status.PodIP, namespace)
			if err != nil {
				continue
			}
			SendMetrics(PrometheusProxyConfig, metrics, pod.Name)
		}
	}
}

func FetchMetricsFromKubeStateMetrics(url string, namespace string) (io.ReadCloser, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s.%s.pod.cluster.local:8080/metrics", strings.ReplaceAll(url, ".", "-"), namespace))
	if err != nil {
		log.Warnf("Unable to fetch Kube State metrics : %s", err.Error())
		return io.NopCloser(bytes.NewBuffer([]byte{})), err
	}
	return resp.Body, nil
}

func FindKubeStateMetricsInNamespace(namespace string) {
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Error(err.Error())
	}

	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, "kube-state-metrics") {
			log.Info(pod.Name)
			metrics, err := FetchMetricsFromKubeStateMetrics(pod.Status.PodIP, namespace)
			if err != nil {
				continue
			}
			SendMetrics(PrometheusProxyConfig, metrics, pod.Name)
		}
	}
}
