# 数据采集scrape模块简介

无论是静态文件 还是通过服务发现，我们都已经解决**监控谁**的问题.那么下面就要拉取监控指标了。数据采集scrape模块主要功能:

- **管理、更新待拉取的目标** : 通过服务发现，`Prometheus`总是会拿到当前最新的监控对象的信息(*例如：拉取`metrics`的地址等*)。`scrape`模块负责拉取这些`target`的指标
- **label与relabeling**: 拉取之前，处理这个指标的初始化标签和`relabel`
- **拉取metrics** : 发起`http(s)`请求,拉取监控指标
- **调用存储**: 将拉取的数据`append`到`storage`模块。`storage`模块负责存储。

注：本章节中代码中会大量出现`Manager`的方法，例如`func (m *Manager) reload()`。若无说明，此`Manager`指的是`Scrape.Manager`

## scrape模块核心逻辑

<img src="./src/scape流程.svg" alt="scrape模块执核心逻辑" style="zoom:200%;" />


### 更新和管理target

由[服务发现的核心逻辑](./discovery_core_logic.md)可知，`discovery`模块通过`syncCh chan map[string][]*targetgroup.Group`向`scrape模块`同步服务发现的结果( *即:`map[string][]*targetgroup.Group`,`key`为`job_name`,`value`为每个`job`对应的`target`地址等信息* )。  
<br>

这个过程主要涉及三个部分：

- 服务发现的地址等信息由`Scrape.Manager`实例进行管控。暂存于`Scrape.Manager`的`targetSets`(*类型:`map[string][]*targetgroup.Group`*)字段中。 

- `scrape模块`启动`update`协程 ( 注：*这个协程执行的函数[func (m *Manager) Run(tsets <-chan map[string][]*targetgroup.Group) error](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/manager.go#L116)* )。`update`协程监听`syncCh chan map[string][]*targetgroup.Group`并且将结果更新到`Scrape.Manager`的`targetSets`字段，然后向`Scrape.Manager`的`triggerReload`字段(类型：`chan struct{}`)发送`reload`信号。

- `reloader`协程( 注：*运行的函数[func (m *Manager) reloader() ](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/manager.go#L139)* ) **定期**地尝试在`Scrape.Manager`的`triggerReload`(类型`chan struct{}` )获取`reload`信号。如果获取到`reload`信号说明，`target`有变化,需要重新加载`target`并启动新的拉取工作。这个`reload`过程由`func (m *Manager) reload()`实现的.


### label与relabeling

在**拉取指标之前**，会先根据配置文件中`relabel_configs`配置项设置标签。这一部分由函数[`func PopulateLabels(lb *labels.Builder, cfg *config.ScrapeConfig, noDefaultPort bool) (res, orig labels.Labels, err error)`](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/target.go#L422)实现的。 具体解析见[Label与ReLabeling](./label_relabeling.md)



### 拉取metrics

`scrape`模块会为每个`job_name`创建独立的`scrapePool`。`scrapePool`负责此`job_name`的拉取指标的工作。每一个`job_name`下，会有一个或多个的`target`。`scrapePool`会为每个具体的`target`创建独立的`scrapeLoop`去拉取指标。


### 调用`storage`模块函数存储指标

在**拉取指标之后**，`scrape模块`会执行[func (sl *scrapeLoop) append(app storage.Appender, b []byte, contentType string, ts time.Time) (total, added, seriesAdded int, err error)](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/scrape.go#L1466)函数，调用`storage`模块函数(例如[app.AppendHistogram](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/scrape.go#L1638)、[app.Append](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/scrape.go#L1643)) `append` 指标数据。
