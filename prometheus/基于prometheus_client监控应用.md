# 基于prometheus client监控应用



`prometheus` 服务器定时从`target`服务收集指标数据。




## 指标(`Metric`)定义

`Prometheus`的指标(`Metric`)被统一定义为： 

```
<metric name>{<label_name_1>=<label_value_1>,<label_name_2>=<label_value_2>,...} 
```

说明：

- 指标名称(`metric name`)：反映被监控的样本,例如`prometheus_http_requests_total`表示 `Prometheus`接收到的`HTTP`请求数量; 指标名称(metric name)命名必须满足如下规则：
  - 指标名称必须有字母、数字、下划线或者冒号组成
  - 不能以数字开头，也就是说必须满足`[a-zA-Z_:][a-zA-Z0-9_:]*`
  - 冒号`:`不得使用于`exporter`
- 标签(`label`)反映样本的特征维度,通过这些维度`Prometheus`可以对样本数据进行过滤，聚合等.标签命名必须满足如下规则：
  - 标签名称必须有字母、数字、下划线或者冒号组成
  - 标签名称不能以数字开头，也就是说必须满足`[a-zA-Z_:][a-zA-Z0-9_:]*`
  - 前缀为`__`标签，是为系统内部使用而预留的。

注：`Prometheus`拉取到的指标(`Metric`)形式都是` <metric name>{<label_name_1>=<label_value_1>,<label_name_2>=<label_value_2>,...} `的。但在存储上，指标名称(`metric name`)将会以`__name__=<metric name>`的形式保存在数据库中的.例如`prometheus_http_requests_total{code="200",handler="/"}`① 会被转换成 `{__name__ = "prometheus_http_requests_total", code="200",handler="/"}`②。所以①、②是同一时序的不同表示而已。



## 指标(`Metric`)类型

`Prometheus`采集到的`Metric`类型有四种：`Counter`、`Gauge`、`Histogram`、`Summary`。  

### Counter(计数器类型)

Counter(计数器类型): 一般用于累计值，**只增不减**，例如记录请求次数、任务完成数、错误发生次数。类比:人生吃饭、喝水的次数  
例如: 接口`/metrics`，状态码为`200`的请求次数

```text
  prometheus_http_requests_total{code="200",handler="/metrics"} 851
```

展示：  
![prometheus_http_requests_total](/Users/tyltr/opencode/readcode/prometheus/src/prometheus_http_requests_total.png "prometheus_http_requests_total")


### Gauge(仪表盘类型)

Gauge(仪表盘类型): 一般的监控指标，波动的指标，**可增可减**，例如cpu使用率，可用内存。类比:每顿吃了几碗饭。 

例如：`go`程序的内存分配情况  

```
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 2.1667616e+07
```

展示  
 ![go_memstats_alloc_bytes](/Users/tyltr/opencode/readcode/prometheus/src/go_memstats_alloc_bytes.png "go_memstats_alloc_bytes")


### c(直方图类型) 

Histogram(直方图类型):表示一段时间范围内对数据进行采样（通常是请求持续时间或响应大小），并能够对其**指定区间**以及总数进行统计，通常它采集的数据展示为直方图。格式`xxxx_bucket{le="<数值>"[,其他标签]} <数值>`，*注：`le`是向上包含的*

例如：

```
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="0.1"} 727
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="0.2"} 727
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="0.4"} 728
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="1"} 728
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="3"} 728
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="8"} 728
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="20"} 728
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="60"} 728
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="120"} 728
prometheus_http_request_duration_seconds_bucket{handler="/metrics",le="+Inf"} 728
```

`prometheus`，调用`/metrics`接口的监控数据。`request_time <= 0.1s`的请求数 727，`request_time <= 0.4s`的请求数 728。  


展示：  

![prometheus_http_request_duration_seconds_bucket](/Users/tyltr/opencode/readcode/prometheus/src/prometheus_http_request_duration_seconds_bucket.png " prometheus_http_request_duration_seconds_bucket")

### Summary(摘要类型)

Summary(摘要类型):表示一段时间范围内对数据进行采样（通常是请求持续时间或响应大小)。格式`xxxx{quantile="< φ>"[,其他标签]} <数值>`，*注：`quantile`百分比*。
例如：

```
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary

go_gc_duration_seconds{quantile="0"} 0.000024251
go_gc_duration_seconds{quantile="0.25"} 0.0003065
go_gc_duration_seconds{quantile="0.5"} 0.000597208
go_gc_duration_seconds{quantile="0.75"} 0. 000893082
go_gc_duration_seconds{quantile="1"} 0.001552459
```

展示：  
![go_gc_duration_seconds](/Users/tyltr/opencode/readcode/prometheus/src/go_gc_duration_seconds.png " go_gc_duration_seconds")



