package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
)

func InitPrometheusCollector() {
	prometheus.MustRegister(requestTotal)
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
