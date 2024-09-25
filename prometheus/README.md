# prometheus

## 目录

- 项目简述与准备
  - [代码版本](项目简述与准备.md#代码版本)
  - [阅读环境](项目简述与准备.md#阅读环境)
  - [手动编译安装与运行](项目简述与准备.md#下载代码)
  - [拉取演示](项目简述与准备.md#拉取演示)
- prometheus简述
  - [prometheus架构](prometheus简述.md)
  - [prometheus功能介绍](prometheus功能介绍.md)
  - [prometheus server模块介绍](prometheus_server模块.md)
  - [prometheus server启动-main函数分析](prometheus_server启动.md)
  - [开发基于prometheus client的target](开发基于prometheus_client的target.md)
- 服务发现
  - [服务发现简述](discovery简述.md)
  - [prometheus监控外部kubernetes集群配置](discovery_k8s_config.md)
  - [prometheus服务发现的核心逻辑](discovery_core_logic.md)
  - [prometheus服务发现的实现](discovery_core_implementation.md)
  - [kubernetes协议的服务发现](discovery_k8s_implementation.md)
  - [新版本DiscoveryManager](discovery_新版本DiscoveryManager.md)
- scrape
  - [数据采集scrape模块简介](scrape_core_logic.md)
  - [数据采集scrape模块代码分析](scrape_work.md)
- Label和Relabeling
  - [Label和Relabeling使用](Label和Relabeling.md)
  - [Label和Relabeling代码解析](Label和Relabeling.md)
- 存储模块解析
  - [存储模块简述](存储模块简述.md)
- TSDB
  - [时序数据](时序数据.md)
  - [时序数据库](时序数据库.md)
  - [tsdb V2说明](tsdbv2说明.md)
  - [tsdb V3原理概述](tsdbV3原理概述.md)
  - [tsdb v3源码解析](tsdbv3源码解析.md)
- PromQL
  - [promql基本语法](promql_syntactic.md)
  - [聚合操作符与函数](promql_aggregation_operators_functions.md)
  - [promql实践与应用](promql_practice.md)
  - [源码分析:promql解析过程](promql_implementation.md)
  - [源码分析:聚合操作符实现](promql_aggregation_implementation.md)
  - [源码分析:promql函数实现](promql_function_implementation.md)
- 规则
  - [规则(todo)](规则.md)
- 告警服务发现
  - todo
- 告警
  - [告警(todo)](告警.md)


# Alert Manager

TODO



# PushGateway

TODO