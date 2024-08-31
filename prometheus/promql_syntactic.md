# PromQL基本语法

`Prometheus` 提供了一种功能表达式语言PromQL，允许用户**实时地**查询和聚合时间序列数据。查询出来的数据可以显示为图形、表格数据。也可以通过`RESTful API`被第三方系统获取。

官方文档: [https://prometheus.io/docs/prometheus/latest/querying/basics/](https://prometheus.io/docs/prometheus/latest/querying/basics/)  

## 基本语法

### 查询结果的数据类型

`PromQL` 查询结果有四种数据类型：  

- **`Instant vector`**（即时向量）每个时间序列，在任意时间点都**只包含一个样本**,例如：`prometheus` 接收到接口`/metrics`的请求数量`prometheus_http_requests_total{handler="/metrics"}` 。在截止到当前时间点，请求数量只有一个样本。如图：
  ![prometheus instant  vector demo](./src/intant_vecor.png)  
  
- **`Range vector`**（范围向量）每个时间序列都包含一系列时间范围内的数据点，即**多个样本**,例如: `prometheus` 接收到接口`/metrics`的最近3分钟之内的请求数量 `prometheus_http_requests_total{handler="/metrics"}[3m]` 请求数量是一组样本。如图：
   ![prometheus range  vector demo](./src/range_vector_demo.png)  

- `Scalar`（标量） 一个简单的浮点值。
- `String` 一个简单的字符串，目前暂未使用。暂时忽略；  
  
### 时序选择器

在`PromQL` 中有两种时序选择器： `Instant Vector Selectors`(即时向量选择器) 和 `Range Vector Selectors`（范围向量选择器）。  

#### Instant Vector Selectors(即时向量选择器)

`Instant Vector Selectors`(即时向量选择器) 对象是一组指定的时间序列，获取每个时序在某个给定的时间戳上的一个样本。官方说明：  

```text
Instant vector selectors allow the selection of a set of time series and a single sample value for each at a given timestamp (point in time).
```

<br>

即时向量选择器由两部分组成：

- metric name：指标名，指定一组时序，必选;
- 标签选择器: 用于过滤时序上的标签，定义于`{}`内，多个过滤条件使用逗号`,` 分割。可选; 标签过滤有四种运算符：
  - `=` 文本完全匹配，用于‘仅包含xxxx’的逻辑
  - `!=` 文本不匹配，用于‘排除xxxx’的逻辑
  - `=~` 选择正则表达式 匹配
  - `!~` 选择正则表达式 不匹配

<br>

最简单形式的即时向量选择器只有`metric name`。 例如： `prometheus_http_requests_total` 表示 `prometheus` 接收到`http`请求数量。如图：

  ![prometheus_http_requests_total_instant_vector](./src/prometheus_http_requests_total_instant_vector.png)  


<br>

带有标签选择器的即时向量选择器。例如:获取`/metrics`接口并且状态码为200的请求数量：  

```text
prometheus_http_requests_total{handler="/metrics",code="200"}
```

<br>

例如:获取`/api/v1/` 为前缀的请求数量：

```text
prometheus_http_requests_total{handler=~ "/api/v1/.+"}
```

#### Range Vector Selectors（范围向量选择器）

`Range Vector Selectors`（范围向量选择器）对象是**一组指定的时间序列**，获取每个时序在给定的**时间范围**上的**一组**样本。范围向量选择器需要在表达式后紧跟一个方括号`[]`来表示选择的时间范围。官方说明：

```text
Range vector literals work like instant vector literals, except that they select a range of samples back from the current instant. Syntactically, a time duration is appended in square brackets ([]) at the end of a vector selector to specify how far back in time values should be fetched for each resulting range vector element. 
```

<br>

支持的时间单位如下，但在生产环境中，一般使用秒级或者分钟级别的数据。

- `ms` - milliseconds
- `s` - seconds
- `m` - minutes
- `h` - hours
- `d` - days - assuming a day always has 24h
- `w` - weeks - assuming a week always has 7d
- `y` - years - assuming a year always has 365d  

<br>

例如:获取`/api/v1/` 为前缀且3分钟内的请求数量

```
prometheus_http_requests_total{handler=~ "/api/v1/.+"}[3m]
```  

## PromQL操作符与关键字

### PromQL操作符

#### 算数运算符

`prometheus`支持算数运算符加(`+`)、减(`-`)、乘(`*`)、除(`/`)、取模(`%)`、乘方(`^`)。只能使用于`instant vector` 和 `Scalar`类型的计算。不能用于`Range vector`（范围向量），只有双方**标签一致**的才能进行计算，即[向量匹配](#向量匹配vector-matching)

##### **示例1**：算数运算符基本使用  

执行`(prometheus_http_requests_total + prometheus_http_requests_total + 1)/2`

![prometheus_http_requests_total_arithmetic_ops_demo](./src/prometheus_http_requests_total_arithmetic_ops_demo.png)  

<br>

##### **示例2**：**错误示例** `Range vector`参与算数运算符  

执行`prometheus_http_requests_total + prometheus_http_requests_total[1m] + 1` ,会报错`parse error: binary expression must contain only scalar and instant vector types`

原因： 算数运算符不能用于`Range vector`（范围向量)

<br>

![prometheus_http_requests_total_arithmetic_ops_demo_error](./src/prometheus_http_requests_total_arithmetic_ops_demo_error.png) 

<br>

##### **示例3** 标签匹配

`instant vector` 与 `instant vector`之间使用算数运算时，会将左侧`instant vector`的标签与右侧`instant vector`的标签进行对比，只有两者标签相同，才能进行算数运算输出结果，即[向量匹配](#向量匹配vector-matching). 
<br>


执行`prometheus_http_requests_total{handler="/api/v1/query"} +  prometheus_http_requests_total{handler="/api/v1/query",code="200"}`  
只能输出` prometheus_http_requests_total{handler="/api/v1/query",code="200",...} `的结果，不可能输出 ` prometheus_http_requests_total{handler="/api/v1/query",code="400",...} `结果  


如图：
![prometheus_http_requests_total_arithmetic_ops_demo_instant_vectors](./src/prometheus_http_requests_total_arithmetic_ops_demo_instant_vectors.png) 



##### **示例4**：不同指标运算

上述示例使用`prometheus_http_requests_total`指标进行演示的，那么不同指标是否可以进行算术运算呢？ 答案当然是可以的，但是必须遵守**标签匹配**的原则，即[向量匹配](#向量匹配vector-matching). 

不同指标分为如下请求：

- 同一类型的不同指标，例如`go_memstats_mallocs_total + prometheus_engine_query_samples_total`  前后都是`counter`类型
- 不同类型的指标,例如 `go_gc_cycles_automatic_gc_cycles_total  + go_sched_goroutines_goroutines`,`go_gc_cycles_automatic_gc_cycles_total`是`counter`类型，后者是`gauge`类型
- `histogram` 和 `summary`只能在本类型之间进行算数运算，因为`histogram`类型中包含特有标签`le`；`summary`类型中包含特有标签`quantile`，无法和其他类型进行标签匹配，即[向量匹配](#向量匹配vector-matching). 


#### 比较运算符

`prometheus`支持比较算符 等于(`==`)、不等于(`!=`)、大于(`>`)、大于等于(`>=`)、小于(`<`)、小于等于(`<=`)。只能使用于`instant vector` 和 `Scalar`类型的计算，不能用于`Range vector`（范围向量）。  

日常工作中，关键字`bool` 经常配合比较运算符使用。`bool`关键字会直接跟在比较运算符之后，如果比较运算为`true`，则返回`1`.否则返回`0`,很适合告警的场景中。在告警场景中,并不需要关心指标值具体是多少，只需关心是否触发告警(即：`true` 或 `false`) 即可。具体应用细节会在[告警](./告警.md)说明。

<br>

##### **示例1:** 比较运算符基本使用

查询出请求量大于50的指标 `prometheus_http_requests_total > 50`  如图  

![prometheus_http_requests_total_greater_50](./src/prometheus_http_requests_total_greater_50.png) 

<br>

##### **示例2:** bool配合比较运算符使用

`prometheus_http_requests_total > bool 50` 查询请求量大于`50`的指标,如果大于`50`，返回`1`；否则返回`0`。 如图所示

![prometheus_http_requests_total_greater_50_bool](./src/prometheus_http_requests_total_greater_50_bool.png)


#### 逻辑运算符

`prometheus`支持逻辑运算符 `and`(交集)、`or`(并集)、`unless`(差集)，只用于`instant vector`之间的运算。为了方便读者理解，本节所有案例使用相同的样本进行说明。
在[http://127.0.0.1:9090/metrics](http://127.0.0.1:9090/metrics)任意选择两个样本，本次选取`go_gc_duration_seconds`、`prometheus_tsdb_wal_fsync_duration_seconds`。  

```text
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.000036501
go_gc_duration_seconds{quantile="0.25"} 0.000103208
go_gc_duration_seconds{quantile="0.5"} 0.000133374
go_gc_duration_seconds{quantile="0.75"} 0.000158749
go_gc_duration_seconds{quantile="1"} 0.0.000524
go_gc_duration_seconds_sum 0.001737125
go_gc_duration_seconds_count 15


# HELP prometheus_tsdb_wal_fsync_duration_seconds Duration of write log fsync.
# TYPE prometheus_tsdb_wal_fsync_duration_seconds summary
prometheus_tsdb_wal_fsync_duration_seconds{quantile="0.5"} NaN
prometheus_tsdb_wal_fsync_duration_seconds{quantile="0.9"} NaN
prometheus_tsdb_wal_fsync_duration_seconds{quantile="0.99"} NaN
prometheus_tsdb_wal_fsync_duration_seconds_sum 0
prometheus_tsdb_wal_fsync_duration_seconds_count 0
```


##### **示例1:** 逻辑运算符-交集基本使用

`A and B`过滤出`A`、`B`的标签相等的指标`A`。 文氏图表示：
![a_and_b](./src/a_and_b.png)  

<br>

上面两个`go_gc_duration_seconds`和 `prometheus_tsdb_wal_fsync_duration_seconds` 只有`quantile="0.5"`一个标签一致。那么执行`go_gc_duration_seconds  and  prometheus_tsdb_wal_fsync_duration_seconds`，则返回`go_gc_duration_seconds{quantile="0.5"} 0.000133583`。 如图： 

![go_gc_heap_allocs_by_size_bytes_bucket and go_gc_heap_frees_by_size_bytes_bucket](./src/go_gc_duration_seconds_and_prometheus_tsdb_wal_fsync_duration_seconds.png)


##### **示例2:** 逻辑运算符-并集基本使用

`A or B` 文氏图表示：
![a_or_b](./src/a_or_b.png)  

<br>

上面两个`go_gc_duration_seconds`和 `prometheus_tsdb_wal_fsync_duration_seconds` 只有`quantile="0.5"`一个标签一致，其他都不一致。那么执行`go_gc_duration_seconds or  prometheus_tsdb_wal_fsync_duration_seconds`，则返回`go_gc_duration_seconds`和 `prometheus_tsdb_wal_fsync_duration_seconds`的并集。 如图： 


![go_gc_heap_allocs_by_size_bytes_bucket or go_gc_heap_frees_by_size_bytes_bucket](./src/go_gc_duration_seconds_or_prometheus_tsdb_wal_fsync_duration_seconds.png)

##### **示例3:** 逻辑运算符-差集基本使用

`A unless B` 文氏图表示：
![a_unless_b](./src/a_unless_b.png)  

上面两个`go_gc_duration_seconds`和 `prometheus_tsdb_wal_fsync_duration_seconds` 只有`quantile="0.5"`一个标签一致，其他都不一致。那么执行`go_gc_duration_seconds unless  prometheus_tsdb_wal_fsync_duration_seconds`，则返回`go_gc_duration_seconds`和 `prometheus_tsdb_wal_fsync_duration_seconds`的差集。 如图： 

![go_gc_heap_allocs_by_size_bytes_bucket unless go_gc_heap_frees_by_size_bytes_bucket](./src/go_gc_duration_seconds_unless_prometheus_tsdb_wal_fsync_duration_seconds.png)

##### **示例4:** 组合使用

TODO

### 向量匹配Vector Matching

#### 关键字


#### 一对一匹配

TODO

#### 多对一和一对多

TODO

### 聚合操作符

 聚合操作符将在[聚合操作符与函数](./promql_aggregation_implementation.md)详细说明


```
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 3.1791e-05
go_gc_duration_seconds{quantile="0.25"} 0.000132875
go_gc_duration_seconds{quantile="0.5"} 0.000149459
go_gc_duration_seconds{quantile="0.75"} 0.000164084
go_gc_duration_seconds{quantile="1"} 0.000354292
go_gc_duration_seconds_sum 0.12827501
go_gc_duration_seconds_count 910
```