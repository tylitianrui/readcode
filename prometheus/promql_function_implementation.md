# 源码分析:promql函数实现

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