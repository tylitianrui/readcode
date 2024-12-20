package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/load"
)

const (
	// HttpApiServerPort = 8520
	// MetricsPort       = 9100

	HttpApiServerPort = 8521
	MetricsPort       = 9101
)

var (
	//HTTP请求总数
	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tyltr_request_total", //  counter
			Help: "HTTP请求总数",
		},
		[]string{"method", "path", "code"}, // labels
	)

	avgLoad = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tyltr_sys_avg_load", //  Gauge
			Help: "系统负载",
		},
		[]string{"min"}, // labels
	)
	respDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tyltr_request_duration",
			Help:    "响应时间",
			Buckets: []float64{2, 5, 8},
		},
		[]string{"path"},
	)
	respDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "pond_temperature_celsius",
			Help:       "The temperature of the frog pond.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},

		[]string{"path"},
	)
)

// 初始化prometheus收集器
func InitPrometheusCollector() {
	prometheus.MustRegister(requestTotal)
	prometheus.MustRegister(avgLoad)
	prometheus.MustRegister(respDuration)
}

// 对外暴漏/metrics接口
func RunPrometheusMetricsApi() {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	// http://127.0.0.1:9100/metrics
	http.ListenAndServe(fmt.Sprintf(":%d", MetricsPort), nil)
}

// 模拟业务处理延迟
func HttpBizLatency() {
	latency := time.Millisecond * time.Duration(rand.Intn(10))
	time.Sleep(latency)
}

// 统计 http请求
func PrometheusRequestTotalMiddleWare(c *gin.Context) {
	meth := c.Request.Method
	path := c.Request.URL.Path
	c.Next()
	code := c.Writer.Status()
	val := fmt.Sprintf("%d", code)
	a := requestTotal.WithLabelValues(meth, path, val)
	a.Add(1)
}

// 统计 http请求
func PrometheusRequestLatency(c *gin.Context) {
	start := time.Now().UnixNano()
	path := c.Request.URL.Path
	c.Next()
	duration := (time.Now().UnixNano() - start) / (1000 * 1000)

	a := respDuration.WithLabelValues(path)
	a.Observe(float64(duration))

}

// http服务，模拟对外正常的业务服务
func RunHttpApiServer() error {
	app := gin.Default()
	// 统计 http请求
	app.Use(PrometheusRequestLatency, PrometheusRequestTotalMiddleWare)

	app.GET("/ping/:id", func(c *gin.Context) {
		HttpBizLatency()
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"id": id,
		})
	})
	return app.Run(fmt.Sprintf(":%d", HttpApiServerPort))
}

// 获取系统负载
func AvgLoad() {
	loadInfo, err := load.Avg()
	if err != nil {
		return
	}
	avgLoad.WithLabelValues("1").Set(loadInfo.Load1)
	avgLoad.WithLabelValues("5").Set(loadInfo.Load5)
	avgLoad.WithLabelValues("15").Set(loadInfo.Load15)
}

// 每秒一次收集负载
func RunAvgLoadCollector() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				AvgLoad()
			}
		}
	}()
}

func main() {
	InitPrometheusCollector()
	go RunPrometheusMetricsApi()
	go RunAvgLoadCollector()

	RunHttpApiServer()
}
