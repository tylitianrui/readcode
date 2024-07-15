# 源码分析:聚合操作符实现

## 关键字定义

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
