
# 2.3 prometheus server中的代码模块


![Prometheus server architecture](src/internal_architecture.svg)

注：此图选自[Prometheus官方图片](https://github.com/prometheus/prometheus/blob/main/documentation/images/internal_architecture.svg)

`prometheus server` 主要的功能模块：

- 服务发现模块：获取`target`的地址等信息，由`Scrape Discovery Manager`进行管理
- `Scrape`模块：拉取`target`的监控指标，由`Scrape Manager`是拉取监控指标的管理者
- 存储模块。`Fanout Storage`是存储层的代理，屏蔽了底层不同存储的实现。无论是本地存储远端存储都有`Fanout Storage`作代理。
- `PromQL`模块，解析执行`PromQL`
- 告警组件服务发现模块，由`Notifier Discovery Manager`进行管理
- 告警模块:`Notifier` 将告警信息发送给`AlertManager`
- 规则模块：主要作用是优化查询规则和触发告警规则，由`Rule Manager`管理规则。
- `TSDB`: `Prometheus`的内置的本地数据库
- *标签功能：代码中不存在此功能专属的模块，代码散在`Scrape`模块中。代码逻辑在获取指标之前，指定打标签或改写标签的计划，按照此计划为获取的指标打标签或改写标签，即`label`与`relabeling` 功能。注：功能嵌入在`Scrape`模块*
