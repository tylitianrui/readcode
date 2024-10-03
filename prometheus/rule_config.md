# 规则配置

`Prometheus` 支持两种类型的规则：记录规则和告警规则。

记录规则： 通过**预先计算经常需要的表达式**或**计算成本高昂的表达式**,将其结果保存为一组新的时间序列；查询时，直接返回这预计算的结果，这种预计算的方式比每次查询时实时计算要快的多。这种方式适用于经常使用到的表达式，例如 grafana里持续展示某个计算结果的场景。

告警规则：



配置时，规则需要定义在独立的`yaml`文件中，然后通过 `rule_files` 字段将规则文件加载到`Prometheus` 中。

例如下面配置内容(*此文件节选自`Prometheus`官方提供的配置文件案例:[documentation/examples/prometheus.yml](https://github.com/prometheus/prometheus/blob/v2.53.0/documentation/examples/prometheus.yml)*)

``````yaml

# my global config
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
  # scrape_timeout is set to the global default (10s).


# Load rules once and periodically evaluate them according to the global 'evaluation_interval'.
rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: "prometheus"

    static_configs:
      - targets: ["localhost:9090"]



``````



说明：

-  `rule_files`   可以通过此字段引入多个规则的配置文件

-  `evaluation_interval`     prometheus会定期地计算是否除非规则，`evaluation_interval`  用于设置定期的时间，默认每1分钟计算一次。

     

## 规则检查



可以使用`Prometheus` 提供的运维工`promtool`   对规则文件进行检查。

**命令**

``````shell
promtool check rules /path/to/example.rules.yml
``````

如果规则是正确的，则返回`SUCCESS: xx rules found`，例如：

```tex
  Checking /path/to/example.rules.yml
  SUCCESS: 1 rules found
```

如果规则是错误的，则返回`FAILED`和相关的错误信息，例如：

``````
Checking /path/to/example.rules.yml
  FAILED:
/path/to/example.rules.yml: yaml: unmarshal errors:
  line 10: cannot unmarshal !!seq into map[string]string
``````



## 规则配置

规则文件的配置语法如下：

``````yaml
groups:
  - name: <string>      # 必须，group的名字，全局唯一
    interval: 15s       # 非必须，默认值为global.evaluation_interval 
    limit: 0            # 非必须，整型，默认值为0
    query_offset: 15s   # 非必须，默认值为global.rule_query_offset 
    rules:              # 规则配置
    - record: code:prometheus_http_requests_total:sum       # 输出的新时序的名字，必须遵守metric的名称规范
      expr: sum by (code) (prometheus_http_requests_total)  # 此为PromQL表达式，prometheus会周期地执行此表达式，产生新的序列。新时序的名称就是上面配置的record
      labels:                                               # 非必须, 在存储新时序之前，新增标签或者改写现有标签
        - label_name1: label_value1
          label_name2: label_value2
``````







## 告警规则 Alerting Rules



