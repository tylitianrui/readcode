# 数据采集scrape模块代码分析

## 关键的数据结构

scrape模块的关键数据结构如下：

- [`scrape.Manager`](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/manager.go#L96) 负责`scrape`模块的核心数据结构，负责管理、更新`target`信息;维护每个`job_name`的`scrapePool`;维护`storage`模块存储实例以便存储指标
- [`scrape.scrapePool`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L64) 管理一组`targets`的拉取任务。`prometheus`会为每个`job_name` 创建一个独立的`scrapePool`,负责管理此`job_name`下所有`target`的指标拉取任务
- [`scrape.scrapeLoop`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L822) 拉取指标的`loop`,周期性地拉取某一target的指标。[scrape.scrapeLoop](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L822)实现了[loop](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L807)接口。
- [`scrape.scrapeCache`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L870) 记录存储过程信息，并且在存储时校验指标的合法性。
  
### `scrape.Manager` 
TODO

### `scrape.scrapePool`

TODO

### `scrape.scrapeLoop`


TODO

### `scrape.scrapeCache`


TODO