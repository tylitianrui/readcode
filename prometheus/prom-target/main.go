package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/load"
)

var (
	// counter  案例
	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tyltr_request_total", //  counter命名  xxxx__total
			Help: "请求总数",
		},
		[]string{"method", "path", "code"}, // labels
	)
	// cpu
	avgLoad1 = prometheus.NewGauge(prometheus.GaugeOpts{

		Name:        "tyltr_sys_avg_load1",
		Help:        "系统负载",
		ConstLabels: map[string]string{},
	})
	avgLoad5 = prometheus.NewGauge(prometheus.GaugeOpts{

		Name:        "tyltr_sys_avg_load5",
		Help:        "系统负载",
		ConstLabels: map[string]string{},
	})
	avgLoad15 = prometheus.NewGauge(prometheus.GaugeOpts{

		Name:        "tyltr_sys_avg_load15",
		Help:        "系统负载",
		ConstLabels: map[string]string{},
	})
)

func InitPrometheusCollector() {
	prometheus.MustRegister(requestTotal)
	prometheus.MustRegister(avgLoad1)
	prometheus.MustRegister(avgLoad5)
	prometheus.MustRegister(avgLoad15)
}

func main() {
	InitPrometheusCollector()
	NewApiServer()
}

func NewApiServer() error {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9100", nil)
	}()
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				CpuUseGauge()
			}

		}

	}()

	app := gin.Default()
	app.Use(PrometheusRequestTotalMiddleWare)

	app.GET("/ping/*name", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	return app.Run(":8520")

}

func PrometheusRequestTotalMiddleWare(c *gin.Context) {
	meth := c.Request.Method
	path := c.Request.URL.Path
	c.Next()
	code := c.Writer.Status()
	requestTotal.WithLabelValues(meth, path, fmt.Sprintf("%d", code)).Add(1)

}

func CpuUseGauge() {
	loadInfo, err := load.Avg()
	if err != nil {
		return
	}
	avgLoad1.Set(loadInfo.Load1)
	avgLoad5.Set(loadInfo.Load5)
	avgLoad15.Set(loadInfo.Load15)

}
