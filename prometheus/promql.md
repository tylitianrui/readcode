# PromQL

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

- ms - milliseconds
- s - seconds
- m - minutes
- h - hours
- d - days - assuming a day always has 24h
- w - weeks - assuming a week always has 7d
- y - years - assuming a year always has 365d  

<br>

例如:获取`/api/v1/` 为前缀且3分钟内的请求数量

```
prometheus_http_requests_total{handler=~ "/api/v1/.+"}[3m]
```  

## PromQL操作符与关键字

### PromQL操作符

#### 算数运算符

`prometheus`支持算数运算符加(+)、减(-)、乘(*)、除(/)、取模(%)、乘方(^)。只能使用于`instant vector` 和 `Scalar`类型的计算。不能用于`Range vector`（范围向量）。  

**示例1**：算数运算符  

执行` (prometheus_http_requests_total + prometheus_http_requests_total + 1)/2  `   

![prometheus_http_requests_total_arithmetic_ops_demo](./src/prometheus_http_requests_total_arithmetic_ops_demo.png)  

<br>

**示例2**：**错误示例** `Range vector`参与算数运算符   
执行` prometheus_http_requests_total + prometheus_http_requests_total[1m] + 1  ` ,会报错`parse error: binary expression must contain only scalar and instant vector types`    

原因： 算数运算符不能用于`Range vector`（范围向量）  
<br>

![prometheus_http_requests_total_arithmetic_ops_demo_error](./src/prometheus_http_requests_total_arithmetic_ops_demo_error.png) 

<br>

**`instant vector` 与 `instant vector`之间使用算数运算**

`instant vector` 与 `instant vector`之间使用算数运算时，会将左侧`instant vector`的标签与右侧`instant vector`的标签进行对比，只有两者标签相同，才能进行算数运算输出结果. 
<br>

**示例3**   

执行`prometheus_http_requests_total{handler="/api/v1/query"} +  prometheus_http_requests_total{handler="/api/v1/query",code="200"} `    
只能输出` prometheus_http_requests_total{handler="/api/v1/query",code="200",...} `的结果，不可能输出 ` prometheus_http_requests_total{handler="/api/v1/query",code="400",...} `结果  

如图：
![prometheus_http_requests_total_arithmetic_ops_demo_instant_vectors](./src/prometheus_http_requests_total_arithmetic_ops_demo_instant_vectors.png) 


#### 比较运算符

`prometheus`支持比较算符 等于(==)、不等于(!=)、大于(>)、大于等于(>=)、小于(<)、小于等于(<=)。只能使用于`instant vector` 和 `Scalar`类型的计算。不能用于`Range vector`（范围向量）。  除了这些比较运算符之外，关键字`bool` 经常配合比较运算符使用。    

`bool`关键字会直接跟在比较运算符之后，如果比较运算为true，则返回1.否则返回0。应用于告警的场景中，在后续【告警】#TODO 进行说明

<br>

**示例:** 比较运算符基本使用   

查询出请求量大于50的指标 ` prometheus_http_requests_total > 50`  如图  

![prometheus_http_requests_total_greater_50](./src/prometheus_http_requests_total_greater_50.png) 

<br>

**示例:** bool配合比较运算符使用   

查询请求量大于50的指标,如果大于50，返回1；否则返回0  ` prometheus_http_requests_total >  bool 50`  如图  

![prometheus_http_requests_total_greater_50_bool](./src/prometheus_http_requests_total_greater_50_bool.png) 


#### 逻辑算符

### 关键字

TODO

#### 分组

使用关键字 `by`进行分组
  
例如: 获取请求`/api/v1/query`的请求总数量

```text
sum(prometheus_http_requests_total{handler="/api/v1/query"})

```

  ![prometheus_http_requests_total_sum](./src/prometheus_http_requests_total_sum.png)  

*注：如果prometheus里暂时没有监控数据，可以手动向prometheus发请求，使prometheus获得http监控数据。例如：*

```shell

curl -X OPTIONS  http://127.0.0.1:9090/api/v1/query  
curl -X PUT http://127.0.0.1:9090/api/v1/query
curl   http://127.0.0.1:9090/api/v1/query  

```

针对上例，获取请求`/api/v1/query`的请求总数量,并且以状态码分组统计  

```shell

sum(prometheus_http_requests_total{handler="/api/v1/query"}) by (code)

```

  ![prometheus_http_requests_total_sum](./src/prometheus_http_requests_total_sum_by_code.png)  



## 聚合操作符与函数介绍

### 聚合操作符

工作中经常使用的 `sum`、`avg`等经常与函数(例如`rate`、`increase`等)混为一谈，其实 `sum`、`avg`并不是函数，而是`prometheus`内置的聚合操作符。聚合操作符和函数的底层实现也不一样。这就是为什么`prometheus`函数的官方文档里找不到他们的原因。  
聚合操作符用于聚合某一个`instant vector`的数据的。  
<br>
聚合操作符目前有12个：  

- **sum** 累加
- **avg**  平均值
- min 最小值
- max 最大值
- stddev 标准差
- stdvar 标准方差
- count  计数
- count_values 对value进行计数
- bottomk  最少的n条时序
- topk  最多的n条时序
- quantile  分位数
- group (all values in the resulting vector are 1)
  
<br>

#### sum

TODO

<br>

#### avg

TODO

### 内置函数

Prometheus提供了其它大量的内置函数，可以对时序数据进行的处理。本部分将列举常见的函数进行说明。其余函数请查看[官方文档](https://prometheus.io/docs/prometheus/latest/querying/functions/)

#### increase

`increase`函数获取区间向量中的第一个和最后一个样本并返回其增长量。  

**使用**：`increase(v range-vector)`  

例如:
- `increase(prometheus_http_requests_total{handler="/metrics"}[1m])`
- `increase(prometheus_tsdb_reloads_total{instance='localhost:9090'}[40s])`

**计算方法**

```
时间区间的最后一个样本 -  时间区间的第一个样本
```
示例 

| 时间   | t1  |t2=t1+15s  |t3=t1+30s  |t4=t1+45s |t5=t1+60s |t6=t1+75s |t7=t1+90s |t8=t1+105s |
| :-----| :---- | :---- |:---- |:---- |:---- |:---- |:---- |:---- |
| `xxxx_total` | v1 | v2  |v3  |v4 |v5| v6  |v7  |v8 |
| `increase(xxxx_total[1m])` | -| -| -| -| `v5-v1`| `v6-v2`| `v7-v3`| `v8-v4`|


#### rate

函数用于计算在指定时间范围内计数器**平均每秒**的增加量。

**使用**：`rate(v range-vector)`

例如:

- `rate(prometheus_http_requests_total{handler="/metrics"}[1m])`
- `rate(prometheus_tsdb_reloads_total{instance='localhost:9090'}[40s])`

**计算方法**

```
(时间区间的最后一个样本 -  时间区间的第一个样本） ➗ 时间区间长度
```
示例 

| 时间   | t1  |t2=t1+15s  |t3=t1+30s  |t4=t1+45s |t5=t1+60s |t6=t1+75s |t7=t1+90s |t8=t1+105s |
| :-----| :---- | :---- |:---- |:---- |:---- |:---- |:---- |:---- |
| `xxxx_total` | v1 | v2  |v3  |v4 |v5| v6  |v7  |v8 |
| `increase(xxxx_total[1m])` | -| -| -| -|`(v5-v1)/60s`| `(v6-v2)/60s`| `(v7-v3)/60s`| `(v8-v4)/60s`|

注： 采样间隔15s


##### 数据外推

TODO


#### irate

`rate`函数计算的是样本的**平均**增长速率，没办法很好的反应瞬时**瞬时**增长率，并且无法避免时间窗口范围内的**长尾问题**。
`irate`通过区间向量中最后两个两本数据来计算区间向量的增长速率，解决时间窗口内的**长尾问题**，并且提高了瞬时变化的灵敏度。
使用：`irate(v range-vector)`

实现计算方法如下：


| 时间   | t1  |t2  |t3  |t4 |t5 |t6 |t7 |t8 |
| :-----| :---- | :---- |:---- |:---- |:---- |:---- |:---- |:---- |
| 采样 | v1 | v2  |v3  |v4 |v5| v6  |v7  |v8 |
| irate | -|`(v2-v1)/(t2-t1)`|`(v3-v2)/(t3-t2)` |`(v4-v3)/(t4-t3)` |`(v5-v4)/(t5-t4)`|`(v6-v5)/(t6-t5)`|`(v7-v6)/(t7-t6)` |`(v8-v7)/(t8-t7)` |

注：  t1 、t2、  ... t8  是连续采样的

<br>

示例:  `prometheus_http_requests_total{handler="/metrics"}`再一段时间范围内的采样与`irate(prometheus_http_requests_total{handler="/metrics"}[1m])`的计算结果

| 时间   | 12:20:00  |12:20:15  |12:20:30  |12:20:45 |12:21:00 |12:21:15 |12:21:30 |12:21:45 |12:22:00 |12:22:15 |12:22:30 |12:22:45 |12:23:00 |
| :-----| :---- | :---- |:---- |:---- |:---- |:---- |:---- |:---- |:---- |:---- |:---- |:---- |:---- |
| `prometheus_http_requests_total{handler="/metrics"}` | `131` | `132`  |`133`  |`134` |`135`| `146`  |`166`  |`170`  |`218` |`238`| `259`  |`269`  |`270` |
| `irate(prometheus_http_requests_total{handler="/metrics"}[1m])` | - | `0.0667` | `0.0667` | `0.0667` | `0.0667` | `0.733` |  `1.33` |  `0.267` |  `3.2` | `1.33` | `1.40` | `0.667` | `0.667` | 

**如图所示**

**`prometheus_http_requests_total{handler="/metrics"}`**

![prometheus_http_requests_total_raw.png](./src/prometheus_http_requests_total_raw.png) 

<br>

**`irate(prometheus_http_requests_total{handler="/metrics"}[1m])`**
![prometheus_http_requests_total_irate.png](./src/prometheus_http_requests_total_irate.png) 




## 最佳实践
TODO

## 代码解析

### 关键字定义

`prometheus`关键字(包含聚合操作符)定于于[`promql/parser/lex.go`](https://github.com/prometheus/prometheus/blob/v2.53.0/promql/parser/lex.go#L101)文件中。

```golang

type ItemType int

// This is a list of all keywords in PromQL.
// When changing this list, make sure to also change
// the maybe_label grammar rule in the generated parser
// to avoid misinterpretation of labels as keywords.
var key = map[string]ItemType{
 // Operators.
 "and":    LAND,
 "or":     LOR,
 "unless": LUNLESS,
 "atan2":  ATAN2,

 // Aggregators.聚合操作符
 "sum":          SUM,
 "avg":          AVG,
 "count":        COUNT,
 "min":          MIN,
 "max":          MAX,
 "group":        GROUP,
 "stddev":       STDDEV,
 "stdvar":       STDVAR,
 "topk":         TOPK,
 "bottomk":      BOTTOMK,
 "count_values": COUNT_VALUES,
 "quantile":     QUANTILE,

 // Keywords.
 "offset":      OFFSET,
 "by":          BY,
 "without":     WITHOUT,
 "on":          ON,
 "ignoring":    IGNORING,
 "group_left":  GROUP_LEFT,
 "group_right": GROUP_RIGHT,
 "bool":        BOOL, 

 // Preprocessors.
 "start": START,
 "end":   END,
}
```

TODO

其他细节见[官方文档](https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators)

### 函数

TODO

`promql`本身只是普通的文本。在解析`promql`语句之后，`Prometheus`会根据用户查询意图执行响应的函数。当前版本`Prometheus`支持75个函数。
`FunctionCalls`(*文件：`promql/functions.go`*) 是一个包含`PromQL`支持的所有函数的`map`：

- `key`:`promql`中的函数名。例如：`rate`、
- `value`:具体执行实现函数

```
var FunctionCalls = map[string]FunctionCall{
 "abs":                funcAbs,
 "absent":             funcAbsent,
 "absent_over_time":   funcAbsentOverTime,
 "acos":               funcAcos,
 "acosh":              funcAcosh,
 "asin":               funcAsin,
 "asinh":              funcAsinh,
 "atan":               funcAtan,
 "atanh":              funcAtanh,
 "avg_over_time":      funcAvgOverTime,
 "ceil":               funcCeil,
 "changes":            funcChanges,
 "clamp":              funcClamp,
  "rate":              funcRate,
    // .......
}
```

其他细节见：[promql函数说明](https://prometheus.io/docs/prometheus/2.53/querying/functions/)

#### rate

TODO

[rate 官方文档](https://prometheus.io/docs/prometheus/2.53/querying/functions/#rate)

## 解析promql语句
