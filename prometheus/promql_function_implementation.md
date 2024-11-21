# 源码分析:promql函数实现

`promql`本身只是普通的文本。在解析`promql`语句之后，`Prometheus`会根据用户查询意图执行响应的函数。当前版本`Prometheus`支持75个函数。
`FunctionCalls`(*文件：`promql/functions.go`*) 是一个包含`PromQL`支持的所有函数的`map`：

- `key`:`promql`中的函数名。例如：`rate`
- `value`:类型`FunctionCall` 具体执行实现函数。

```go
// FunctionCalls is a list of all functions supported by PromQL, including their types.
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
	"clamp_max":          funcClampMax,
	"clamp_min":          funcClampMin,
	"cos":                funcCos,
	"cosh":               funcCosh,
	"count_over_time":    funcCountOverTime,
	"days_in_month":      funcDaysInMonth,
	"day_of_month":       funcDayOfMonth,
	"day_of_week":        funcDayOfWeek,
	"day_of_year":        funcDayOfYear,
	"deg":                funcDeg,
	"delta":              funcDelta,
	"deriv":              funcDeriv,
	"exp":                funcExp,
	"floor":              funcFloor,
	"histogram_avg":      funcHistogramAvg,
	"histogram_count":    funcHistogramCount,
	"histogram_fraction": funcHistogramFraction,
	"histogram_quantile": funcHistogramQuantile,
	"histogram_sum":      funcHistogramSum,
	"histogram_stddev":   funcHistogramStdDev,
	"histogram_stdvar":   funcHistogramStdVar,
	"holt_winters":       funcHoltWinters,
	"hour":               funcHour,
	"idelta":             funcIdelta,
	"increase":           funcIncrease,
	"irate":              funcIrate,
	"label_replace":      funcLabelReplace,
	"label_join":         funcLabelJoin,
	"ln":                 funcLn,
	"log10":              funcLog10,
	"log2":               funcLog2,
	"last_over_time":     funcLastOverTime,
	"mad_over_time":      funcMadOverTime,
	"max_over_time":      funcMaxOverTime,
	"min_over_time":      funcMinOverTime,
	"minute":             funcMinute,
	"month":              funcMonth,
	"pi":                 funcPi,
	"predict_linear":     funcPredictLinear,
	"present_over_time":  funcPresentOverTime,
	"quantile_over_time": funcQuantileOverTime,
	"rad":                funcRad,
	"rate":               funcRate,
	"resets":             funcResets,
	"round":              funcRound,
	"scalar":             funcScalar,
	"sgn":                funcSgn,
	"sin":                funcSin,
	"sinh":               funcSinh,
	"sort":               funcSort,
	"sort_desc":          funcSortDesc,
	"sort_by_label":      funcSortByLabel,
	"sort_by_label_desc": funcSortByLabelDesc,
	"sqrt":               funcSqrt,
	"stddev_over_time":   funcStddevOverTime,
	"stdvar_over_time":   funcStdvarOverTime,
	"sum_over_time":      funcSumOverTime,
	"tan":                funcTan,
	"tanh":               funcTanh,
	"time":               funcTime,
	"timestamp":          funcTimestamp,
	"vector":             funcVector,
	"year":               funcYear,
}
```

解析常用的函数 `rate`, `irate`,`histogram_quantile`


## histogram_quantile