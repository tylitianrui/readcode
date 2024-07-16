# prometheus

## 目录
- [项目简述与准备](项目简述与准备.md)
  - [代码版本](项目简述与准备.md#代码版本)
  - [阅读环境](项目简述与准备.md#阅读环境)
  - [下载代码](项目简述与准备.md#下载代码)
  - [手动编译安装与运行](项目简述与准备.md#下载代码)
  - [开发应用接入prometheus监控](项目简述与准备.md#开发应用接入prometheus监控)
  - [拉取演示](项目简述与准备.md#拉取演示)
- [prometheus简述](prometheus简述.md)
  - [prometheus架构](prometheus简述.md#11-架构)
  - [prometheus功能介绍](prometheus简述.md#12-功能介绍)
    - [服务发现sd](prometheus简述.md#121-服务发现)
    - [数据采集scrape](prometheus简述.md#122-数据采集)
    - [数据处理](prometheus简述.md#123-数据处理)
      - 数据类型(todo)
      - 四种Metric类型
    - [数据处理](prometheus简述.md#123-数据处理)
    - [数据存储storage](prometheus简述.md#124-数据存储)
    - [查询query](prometheus简述.md#125-查询)
    - [告警alert](prometheus简述.md#125-告警)
- [prometheus server简述](prometheus_server简述.md)
- [服务发现](服务发现.md)
  - [prometheus监控外部kubernetes集群配置](discovery_k8s_config.md)
  - [prometheus服务发现的核心逻辑](discovery_core_logic.md)
  - [prometheus服务发现的实现](discovery_core_implementation.md)
  - [kubernetes协议的服务发现](discovery_k8s_implementation.md)
- [scrape](scrape.md)
- [relabel](relabel.md)
- [storage(todo)](storage.md)
- [tsdb(todo)](tsdb.md)
- [lables(todo)](lables.md)
- promql
  - [promql基本语法](promql_syntactic.md)
  - [聚合操作符与函数](aggregation_operators_functions.md)
  - [promql实践与应用](promql_practice.md)
  - [源码分析:promql解析过程](promql_implementation.md)
  - [源码分析:聚合操作符实现](promql_aggregation_implementation.md)
  - [源码分析:promql函数实现](promql_function_implementation.md)
- [规则(todo)](规则.md)
- [告警(todo)](告警.md)
- [指标类型结构(todo)](指标类型.md)
- [代理模式(todo)](代理模式.md)


# alertmanager
TODO
