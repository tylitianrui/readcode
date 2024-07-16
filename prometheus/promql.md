# PromQL

`Prometheus` 提供了一种功能表达式语言PromQL，允许用户**实时地**查询和聚合时间序列数据。查询出来的数据可以显示为图形、表格数据。也可以通过`RESTful API`被第三方系统获取。

官方文档: [https://prometheus.io/docs/prometheus/latest/querying/basics/](https://prometheus.io/docs/prometheus/latest/querying/basics/)  

**



## 聚合操作符与函数介绍









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
