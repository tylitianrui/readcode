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

	avgLoad = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tyltr_sys_avg_load",
			Help: "系统负载",
		},
		[]string{"min"},
	)
)

func InitPrometheusCollector() {
	prometheus.MustRegister(requestTotal)
	prometheus.MustRegister(avgLoad)
}

func main() {
	InitPrometheusCollector()
	NewApiServer()
}

type Wrapper struct {
}

func (h *Wrapper) printReq(r *http.Request) {
	fmt.Println("\n\r\n\r\n\r")
	fmt.Println("method:", r.Method)
	fmt.Println("url:", r.Host+r.URL.String())
	if len(r.Header) > 0 {
		fmt.Println("-----Query start------")
	}
	for k, v := range r.Header {
		if len(v) == 1 {
			fmt.Println(k, "=", v[0])
		} else {
			fmt.Println(k, "=", v)
		}

	}
	if len(r.Header) > 0 {
		fmt.Println("-----Query end------")
	}

	q := r.URL.Query()
	if len(q) > 0 {
		fmt.Println("-----Query start------")
	}

	for k, v := range q {
		fmt.Println("k:", k, "val:", v)
	}
	if len(q) > 0 {
		fmt.Println("-----Query end------")
	}
}

func (h *Wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.printReq(r) //  打印请求参数
	promhttp.Handler().ServeHTTP(w, r)
}

func NewApiServer() error {
	go func() {
		http.Handle("/metrics", &Wrapper{})
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
	avgLoad.WithLabelValues("1").Set(loadInfo.Load1)
	avgLoad.WithLabelValues("5").Set(loadInfo.Load5)
	avgLoad.WithLabelValues("15").Set(loadInfo.Load15)

}
