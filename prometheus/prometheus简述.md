# prometheus简述


## prometheus监控系统的架构

![prometheus architecture](./src/architecture.png "prometheus architecture")


## 核心组件

### Prometheus Server
  
`Prometheus Server`是`Prometheus`体系里最核心的组件。负责**服务发现**、**拉取监控指标**、**存储和查询指标数据**、**触发告警**,并提供`API`以供访问。
`Prometheus Server`也是一个时序数据库(`TSDB`)，将采集到的监控数据按照时间序列的方式存储在本地磁盘当中。
  
### Prometheus Web UI
  
内置的`WEB UI`界面，通过这个UI可以直接通过`PromQL`实现数据的查询以及可视化。虽然内置ui界面，但`UI`过于简单，生产环境中一般都不会使用内置`UI`界面，而是使用`grafana`等图形化界面，文档 [https://prometheus.io/docs/visualization/grafana/](https://prometheus.io/docs/visualization/grafana/)。在后续文档中，会介绍`grafana`的应用。

### Exporter

`Prometheus`

`Prometheus server`原理是通过 `HTTP` 协议周期性抓取被监控组件的状态，输出这些被监控的组件的 `Http` 接口为 `Exporter`。
 `Prometheus server`通过轮询或指定的抓取器从`Exporter`提供的`Endpoint`端点中提取数据。
  
### AlertManager
  
`AlertManager` 是告警管理器组件。实践中，会基于`PromQL`创建告警规则,如果满足PromQL定义的规则，就会产生一条告警;`Prometheus server`会将此告警`push`到`AlertManager`;由`AlertManager`完成告警管理(告警聚合、静默、告警等级、告警渠道等)

### PushGateway

`Prometheus`数据采集基于`Pull`模型进行设计，在网络环境必须要让`Prometheus Server`能够直接与`Exporter`进行通信，当这种网络需求无法直接满足时，就可以利用`PushGateway`来进行中转。营业将内部网络的监控数据主动`push`到`PushGateway`当中，`Prometheus Server`则可以采用同样`Pull`的方式从`PushGateway`中获取到监控数据。

  - 应用场景：
    - `Prometheus`和`target` 由于某些原因网络不能互通，需要经由`Pushgateway`代理
    - 短生命周期的任务。因为`Prometheus`是定期`pull`任务的监控信息，也就是有时间间隔；短生命周期的任务，可能采集的时候就已经退出了，那么监控信息都会消失，所以采用主动`push`的方式存入`PushGateway`。例如`kubernetes jobs` 或 `Cronjobs`中收集自定义指标
