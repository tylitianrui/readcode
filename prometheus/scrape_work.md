# 数据采集scrape模块代码分析

## 关键的数据结构

scrape模块的关键数据结构如下：

- [`scrape.Manager`](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/manager.go#L96) 负责`scrape`模块的核心数据结构，负责管理、更新`target`信息;维护每个`job_name`的`scrapePool`;维护`storage`模块存储实例以便存储指标
- [`scrape.scrapePool`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L64) 管理一组`targets`的拉取任务。`prometheus`会为每个`job_name` 创建一个独立的`scrapePool`,负责管理此`job_name`下所有`target`的指标拉取任务。`scrapePool`会为此`job_name`下的所有`target`创建独立的`scrapeLoop`拉取指标
- [`scrape.scrapeLoop`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L822) 拉取指标的`loop`,周期性地拉取某一target的指标。[scrape.scrapeLoop](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L822)实现了[loop](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L807)接口。
- [`scrape.scrapeCache`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L870) 记录存储过程信息，并且在存储时校验指标的合法性。
  
### `scrape.Manager` 

文件: [`scrape/manager.go`](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/manager.go#L96)

```golang
// Manager maintains a set of scrape pools and manages start/stop cycles
// when receiving new target groups from the discovery manager.
// scrape.Manager维护一组scrapePool,scrapePool负责拉取监控指标等工作
// 当通过 discover manager 获取到当前最新的抓取目标的时，scrape.Manager热更新最新的监控目标 并管理scrapePool循环的的启动、关闭  

type Manager struct {
	opts      *Options
	logger    log.Logger
	append    storage.Appendable                      // 存储
	graceShut chan struct{}                           // 关闭信号

	offsetSeed    uint64     
	mtxScrape     sync.Mutex 
	scrapeConfigs map[string]*config.ScrapeConfig    // prometheus.yml配置文件中scrape_configs模块信息: 拉取的target配置的初始值信息,key为job_name
	scrapePools   map[string]*scrapePool             // 存储了一组拉取指标的实际执行者
	targetSets    map[string][]*targetgroup.Group    // target更新要拉取的具体target,key为job_name
	buffers       *pool.Pool

	triggerReload chan struct{}                      // 传递reload信号的channel，通过监听此channel进行reload操作

	metrics *scrapeMetrics                           // 对scrape模块监控指标
}
```  

**主要字段**：
| 字段名   | 类型    |说明 | 
| :-----| :---- | :---- |
| `scrapeConfigs`  |`map[string]*config.ScrapeConfig` | `prometheus.yml`配置文件中`scrape_configs`部分的信息。<br/>  `map`的`key`是`prometheus.yml`配置文件中的`job_name`<br/>  `map`的`value`是对应的配置内容  |
| `scrapePools`   |`map[string]*scrapePool` | `prometheus`会为每个`job_name`创建一个独立的`scrapePool`,此字段存储各个`scrapePool`实例<br/>  `map`的`key`是`job_name`<br/>  `map`的`value`类型是`scrapePool` |
| `targetSets`   |`map[string][]*targetgroup.Group` | 服务发现模块会将当前最新的监控对象封装成`map[string][]*targetgroup.Group`，通过`channel`(注：`channel`的类型`chan map[string][]*targetgroup.Group`)发送给`scrape.Manager`。<br/>   `scrape.Manager`会把接收到的信息暂存在`targetSets`字段。<br/>  `map`的`key`是`job_name`,<br/>  `map`的`value`就是对应的监控对象信息 |
| `triggerReload`  |`chan struct{}`  | 用于传递热更新信号，<br/>  `scrape.Manager`将接收到的信息暂存在`targetSets`字段后，会向`triggerReload`发送更新信号。`scrape.Manager`的`reloader`方法接收到更新信号后，调用更新操作。 |

### `scrape.scrapePool`



### `scrape.scrapeLoop`


TODO

### `scrape.scrapeCache`


TODO


