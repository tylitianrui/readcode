
# prometheus server模块


![Prometheus server architecture](src/internal_architecture.svg)

注：此图选自[Prometheus官方图片](https://github.com/prometheus/prometheus/blob/main/documentation/images/internal_architecture.svg)

prometheus server 主要的功能模块：

- `target`服务发现模块，由`Scrape Discovery Manager`进行管理
- 拉取监控指标模块，`Scrape Manager`是拉取监控指标的实际工作者
- 规则模块，由`Rule Manager`管理规则
- `Storage` 存储模块。`Fanout Storage`是存储层的代理，屏蔽了底层不同存储的实现。
- 告警组件服务发现模块，由`Notifier Discovery Manager`进行管理
- 告警模块:`Notifier` 将告警信息发送给`AlertManager`
- `PromQL`组件

## Scrape Discovery Manager

在`prometheus`中`target`的服务发现由`Scrape Discovery Manager`进行统一管理的。`Scrape Discovery Manager`会不断获取`targets`最新的服务地址等信息；并且`Scrape Discovery Manager`将最新的`target`地址等信息封装成`targetgroup.Group`，并且通过`chan map[string][]*targetgroup.Group`发送给`Scrape Manager`。`Scrape Manager`根据服务发现的结果拉取`target`的指标。 

`chan map[string][]*targetgroup.Group`说明：

- `map`的`key`:   类型`string` 对应配置文件中`job`名称; 
  
- `map`的`value`  类型`[]*targetgroup.Group` 表示此`job`所包含的一系列`targetgroup.Group`。在在`prometheus`中，`target`的信息会由一组标签进行描述，节选部分标签例下:

```text
"__address__": "192.168.0.107:6443",
"__meta_kubernetes_endpoint_port_name": "https",
"__meta_kubernetes_endpoint_port_protocol": "TCP", 
"__meta_kubernetes_endpoint_ready": "true"
```
`targetgroup.Group`本质就是**一组描述服务地址的标签**和其他**公共标签**的集合体.


## Scrape Manager

`Scrape Manager`负责的工作：

- 拉取target的监控指标
- 将获取到的指标样本发送个存储模块

`Scrape Manager`通过`chan map[string][]*targetgroup.Group`获取到服务信息。


