# 数据采集scrape模块代码分析

## 关键的数据结构

scrape模块的关键数据结构如下：

- [`scrape.Manager`](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/manager.go#L96) 负责`scrape`模块的核心数据结构，负责管理、更新`target`信息;维护每个`job_name`的`scrapePool`;维护`storage`模块存储实例以便存储指标
- [`scrape.scrapePool`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L64) 管理一组`targets`的拉取任务。`prometheus`会为每个`job_name` 创建一个独立的`scrapePool`,负责管理此`job_name`下所有`target`的指标拉取任务。`scrapePool`会为此`job_name`下的所有`target`创建独立的`scrapeLoop`拉取指标
- [`scrape.scrapeLoop`](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L822) 拉取指标的`loop`,周期性地拉取某一target的指标。[scrape.scrapeLoop](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L822)实现了[loop](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L807)接口。
  
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

#### `Manager.targetSets`

`Manager.targetSets`字段类型`map[string][]*targetgroup.Group`,暂存了服务发现的结果，那么“这个结果”都是数据呢？

<br>

`targetgroup.Group`定义：

```golang
type Group struct {
    Targets []model.LabelSet  // model.LabelSet 本质是map[string]string
    Labels model.LabelSet     // model.LabelSet 本质是map[string]string
    Source string
}
```

可见，`targetgroup.Group`底层是`map[string]string`的类型,本质上就是**一组kv的标签**

我们以静态文件配置为例进行说明，`prometheus.yml`中`scrape_configs`部分的配置，如下：  

```yaml
scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
  - job_name: "job-0"
    metrics_path: '/metrics'
    scheme : 'http'
    static_configs:
      - targets: ["127.0.0.1:8520","192.168.0.103:8520"]
```  

<br/>

那么`Scrape.Manager.targetSets`对应的值为

```json
{
    "prometheus": [
        {
            "Targets": [
                {
                    "__address__": "localhost:9090"
                }
            ],
            "Labels": null,
            "Source": "0"
        }
    ],
    "job-0": [
        {
            "Targets": [
                {
                    "__address__": "127.0.0.1:8520"
                },
                {
                    "__address__": "192.168.0.103:8520"
                }
            ],
            "Labels": null,
            "Source": "0"
        }
    ]
}

```

### `scrape.scrapePool`

`scrape.Manager`会为每个`job`维护独立的`scrapePool`,`scrapePool`为每个`target`创建独立的`loop`定期向`target`发送`http`请求获取指标。这些`loop`共享`scrapePool`实例的`http client`。  

如果`target`发生更新，`scrapePool`会为新的`target`创建`loop`,关闭失效的`target`的`loop`

文件: [`scrape/scrape.go`](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/scrape.go#L63)

```go
// scrapePool manages scrapes for sets of targets.

type scrapePool struct {
    appendable storage.Appendable       // 存储,此接口定义了存储的行为
    logger     log.Logger
    cancel     context.CancelFunc
    httpOpts   []config_util.HTTPClientOption

    // mtx must not be taken after targetMtx.
    mtx    sync.Mutex
    config *config.ScrapeConfig       // 抓取的配置
    client *http.Client               // http client,用于pull指标时 发起http请求
    loops  map[uint64]loop

    targetMtx sync.Mutex
    // activeTargets and loops must always be synchronized to have the same
    // set of hashes.
    activeTargets       map[uint64]*Target     // 抓取的目标endpoint等信息
    droppedTargets      []*Target // Subject to KeepDroppedTargets limit.
    droppedTargetsCount int       // Count of all dropped targets.

    // Constructor for new scrape loops. This is settable for testing convenience.
    newLoop func(scrapeLoopOptions) loop

    noDefaultPort bool

    metrics *scrapeMetrics       // 监控指标
}
```

**主要字段**：

| 字段名   | 类型    |说明 |
| :-----| :---- | :---- |
| `appendable`  |`storage.Appendable` | 存储模块的实例 |
| `loops` |`map[uint64]loop` | 一组`loop`,注:每个`target`都有独立的`loop`，`loop`是接口类型，`scrape.scrapeLoop`结构体实现了`loop`接口，`scrape.scrapeLoop`才是`loop`的实体|
| `activeTargets`  |`map[uint64]*Target` | 需要抓取指标的一组`target`的信息 <br/> 注:`loops`和`activeTargets`是存在对应关系的，两者都是`kv`结构,`key`都是`target`的哈希值.`loops`和`activeTargets`就是通过哈希值进行关联的|
| `droppedTargets`|`[]*Target`| 需要丢弃的目标,即不需要抓取的`target`|
| `client`   |`*http.Client` | `loop`定期向`target`发送`http`请求获取指标。`client`就是此`http`的客户端|

**思考题**  

1. `activeTargets`、`droppedTargets`都是`Target`集合，为何`activeTargets`以map进行组织，`droppedTargets`选择切片呢？  
TODO

### `scrape.scrapeLoop`

`scrape.scrapeLoop` 实现了 `scrape.loop`接口，是拉取指标的`loop`实体。

文件: [`scrape/scrape.go`](https://github.com/prometheus/prometheus/blob/v2.53.0/scrape/scrape.go#L812)

```go
type scrapeLoop struct {
    scraper                  scraper
    l                        log.Logger
    cache                    *scrapeCache
    lastScrapeSize           int
    buffers                  *pool.Pool
    offsetSeed               uint64
    honorTimestamps          bool
    trackTimestampsStaleness bool
    enableCompression        bool
    forcedErr                error
    forcedErrMtx             sync.Mutex
    sampleLimit              int
    bucketLimit              int
    maxSchema                int32
    labelLimits              *labelLimits
    interval                 time.Duration
    timeout                  time.Duration
    scrapeClassicHistograms  bool

    // Feature flagged options.
    enableNativeHistogramIngestion bool
    enableCTZeroIngestion          bool

    appender            func(ctx context.Context) storage.Appender
    symbolTable         *labels.SymbolTable
    sampleMutator       labelsMutator
    reportSampleMutator labelsMutator

    parentCtx   context.Context
    appenderCtx context.Context
    ctx         context.Context
    cancel      func()
    stopped     chan struct{}

    disabledEndOfRunStalenessMarkers bool

    reportExtraMetrics  bool
    appendMetadataToWAL bool

    metrics *scrapeMetrics

    skipOffsetting bool // For testability.
}
```

**主要字段**：

| 字段名   | 类型    |说明 |
| :-----| :---- | :---- |
| `scraper`  |`scraper` | 封装了http请求target，获取指标的过程 |
| `cache` |`*scrapeCache` | |
| `appender`  |`func(ctx context.Context) storage.Appender` |获取存储模块的存储实例，用于指标存储|

## HOW TO WORK

### 创建`ScrapeManager`与监听服务发现

**创建`ScrapeManager`实例**  

<br>

```go
    scrapeManager, err := scrape.NewManager(
        &cfg.scrape,     // 配置文件中的配置信息
        log.With(logger, "component", "scrape manager"),
        fanoutStorage,   // 存储模块的代理，屏蔽底层的存储实现
        prometheus.DefaultRegisterer,
    )
    if err != nil {
        level.Error(logger).Log("msg", "failed to create a scrape manager", "err", err)
        os.Exit(1)
    }
```

**监听服务发现**

<br>

```go
    {
        // Scrape manager.
        g.Add(
            func() error {
                // When the scrape manager receives a new targets list
                // it needs to read a valid config for each job.
                // It depends on the config being in sync with the discovery manager so
                // we wait until the config is fully loaded.
                <-reloadReady.C  

                // 监听服务发现
                err := scrapeManager.Run(discoveryManagerScrape.SyncCh())    
                level.Info(logger).Log("msg", "Scrape manager stopped")
                return err
            },

            func(err error) {
                // Scrape manager needs to be stopped before closing the local TSDB
                // so that it doesn't try to write samples to a closed storage.
                // We should also wait for rule manager to be fully stopped to ensure
                // we don't trigger any false positive alerts for rules using absent().
                level.Info(logger).Log("msg", "Stopping scrape manager...")
                scrapeManager.Stop()
            },
        )
    }
```

### 更新`targets`

图例  

![scrape流程update_targets](./src/scrape流程update_targets.drawio.svg)

更新`targets`流程说明：

`Manager.Run`接收到服务发现的结果(`map[string][]*targetgroup.Group`)后：

- `m.updateTsets(ts)`：把接收到的信息(`targetgroup.Group`)暂存在`m.targetSets`字段
- `m.triggerReload <- struct{}{}`: `m.triggerReload`发送`reload`信号
- `go m.reloader()`启动的协程`reloader`，定期(默认5s)轮询 `m.triggerReload`。如果获取到`reload`信号，执行`Manager.reload()` 方法
  
代码解析：

```go
// Run receives and saves target set updates and triggers the scraping loops reloading.
// Reloading happens in the background so that it doesn't block receiving targets updates.

func (m *Manager) Run(tsets <-chan map[string][]*targetgroup.Group) error {
    go m.reloader() // 协程启动reloader， 监听更新信息
    // 循环
    for {
        select {
        case ts := <-tsets:   // 在chan tsets 获取到当前最新的拉取对象的信息, chan tsets的send端一般是服务发现组件
            m.updateTsets(ts) // 更新targets,将 m.targetSets 设置为ts

            select {
            case m.triggerReload <- struct{}{}:  // 发生reload信号
            default:
            }

        case <-m.graceShut:  //  关闭信号
            return nil
        }
    }
}


// 将 m.targetSets 设置为ts
func (m *Manager) updateTsets(tsets map[string][]*targetgroup.Group) {
    m.mtxScrape.Lock()
    m.targetSets = tsets
    m.mtxScrape.Unlock()
}


// 监听reload信号 触发更新操作
func (m *Manager) reloader() {
    reloadIntervalDuration := m.opts.DiscoveryReloadInterval
    if reloadIntervalDuration < model.Duration(5*time.Second) {
        reloadIntervalDuration = model.Duration(5 * time.Second)
    }

    ticker := time.NewTicker(time.Duration(reloadIntervalDuration))

    defer ticker.Stop()

    for {
        select {
        case <-m.graceShut:
            return
        case <-ticker.C:  // 定期轮训 m.triggerReload
            select {
            case <-m.triggerReload: // 监听到reload信号，执行reload操作
                m.reload()          // 实际上加载targets的操作
            case <-m.graceShut:
                return
            }
        }
    }
}
```

### 为job创建ScrapePool

`func (m *Manager) reload()`为每个`job_name`创建独立的`ScrapePool`，并将`ScrapePool`实例存储于`Manager.scrapePools`字段中。如图

![scape流程_create_scrapepool](./src/scape流程_create_scrapepool.drawio.svg)

<br>

代码解析  

```go
func (m *Manager) reload() {
    m.mtxScrape.Lock()
    var wg sync.WaitGroup
    // m.targetSets 暂存的是当前最新的抓取目标，是在 Manager.Run--> Manager.updateTsets(ts) 中进行设置的   
    // 遍历m.targetSets为每个job创建 scrapePool
    for setName, groups := range m.targetSets {
        // check 是否存在jop_name的scrapePools，如果不存在，则创建
        if _, ok := m.scrapePools[setName]; !ok {
            // 配置文件中是否存在此job
            scrapeConfig, ok := m.scrapeConfigs[setName]
            if !ok {
                level.Error(m.logger).Log("msg", "error reloading target set", "err", "invalid config id:"+setName)
                continue
            }
            // 为每个job_name创建scrapePool实例
            m.metrics.targetScrapePools.Inc()  // 创建scrapePool实例,监控指标 +1 
            sp, err := newScrapePool(scrapeConfig, m.append, m.offsetSeed, log.With(m.logger, "scrape_pool", setName), m.buffers, m.opts, m.metrics)
            if err != nil {
                m.metrics.targetScrapePoolsFailed.Inc() // 创建失败,监控指标 +1 
                level.Error(m.logger).Log("msg", "error creating new scrape pool", "err", err, "scrape_pool", setName)
                continue
            }
            m.scrapePools[setName] = sp
        }

        // 启动协程，向scrapePool同步最新的Target Group
        // sp.Sync(groups)  将 Target Group 转换为实际的抓取目标Target，
        // 同步当前运行的 scraper 和结果集，返回全部抓取和丢弃的目标。
        wg.Add(1)
        // Run the sync in parallel as these take a while and at high load can't catch up.
        go func(sp *scrapePool, groups []*targetgroup.Group) {
            sp.Sync(groups)
            wg.Done()
        }(m.scrapePools[setName], groups)

    }
    m.mtxScrape.Unlock()
    wg.Wait()
}

```

### 标签处理

在之前已经获取到服务发现的所有信息了，下面就需要完成一件重要的工作：

- 将服务发现的信息转换成具体的`target`地址(*见:②*),并且将最新的`target`地址同步给拉取指标的执行者(*见:④*)
- 用户自定义标签处理和`relabel`(*见:③*)

**代码调用链**  

①`func (sp *scrapePool) Sync(tgs []*targetgroup.Group)`  -->  ②`func TargetsFromGroup(tg *targetgroup.Group, cfg *config.ScrapeConfig, noDefaultPort bool, targets []*Target, lb *labels.Builder) ([]*Target, []error)`   -->  ③`func PopulateLabels(lb *labels.Builder, cfg *config.ScrapeConfig, noDefaultPort bool) (res, orig labels.Labels, err error)`  --> ④`func (sp *scrapePool) sync(targets []*Target)`

<br/>

**代码解析**  

```golang
// Sync converts target groups into actual scrape targets and synchronizes
// the currently running scraper with the resulting set and returns all scraped and dropped targets.
func (sp *scrapePool) Sync(tgs []*targetgroup.Group) {
    sp.mtx.Lock()
    defer sp.mtx.Unlock()
    start := time.Now() 
    sp.targetMtx.Lock()
    var all []*Target
    var targets []*Target
    lb := labels.NewBuilderWithSymbolTable(sp.symbolTable)
    sp.droppedTargets = []*Target{}
    sp.droppedTargetsCount = 0
    // 
    for _, tg := range tgs {
        targets, failures := TargetsFromGroup(tg, sp.config, sp.noDefaultPort, targets, lb)
        for _, err := range failures {
            level.Error(sp.logger).Log("msg", "Creating target failed", "err", err)
        }
        sp.metrics.targetSyncFailed.WithLabelValues(sp.config.JobName).Add(float64(len(failures)))
        for _, t := range targets {
            // Replicate .Labels().IsEmpty() with a loop here to avoid generating garbage.
            nonEmpty := false
            t.LabelsRange(func(l labels.Label) { nonEmpty = true })
            switch {
            case nonEmpty:
                all = append(all, t)
            case !t.discoveredLabels.IsEmpty():
                if sp.config.KeepDroppedTargets == 0 || uint(len(sp.droppedTargets)) < sp.config.KeepDroppedTargets {
                    sp.droppedTargets = append(sp.droppedTargets, t)
                }
                sp.droppedTargetsCount++
            }
        }
    }
    sp.targetMtx.Unlock()
    sp.sync(all)    
    sp.metrics.targetSyncIntervalLength.WithLabelValues(sp.config.JobName).Observe(
        time.Since(start).Seconds(),
    )
    sp.metrics.targetScrapePoolSyncsCounter.WithLabelValues(sp.config.JobName).Inc()
}
```

### scrape拉取指标

TODO

### 指标存储

TODO
