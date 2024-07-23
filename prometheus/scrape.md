# 数据采集scrape模块

无论是静态文件 还是通过服务发现，我们都已经解决**监控谁**的问题.那么下面就要拉取监控指标了。数据采集scrape模块主要功能:

- **管理、更新待拉取的目标** : 通过服务发现，`Prometheus`总是会拿到当前最新的监控对象的信息(*例如：拉取`metrics`的地址等*)。`scrape`模块负责拉取这些`target`的指标
- **label与relabeling**: 拉取之前，处理这个指标的初始化标签和`relabel`
- **拉取metrics** : 发起`http(s)`请求,拉取监控指标
- **调用存储**: 将拉取的数据`append`到`storage`模块。`storage`模块负责存储。

注：本章节中代码中会大量出现`Manager`的方法，例如`func (m *Manager) reload()`。若无说明，此`Manager`指的是`Scrape.Manager`

## scrape模块核心逻辑

![scrape模块执核心逻辑](./src/scape流程.svg)


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



<!-- 

### 更新targets   

`Scrape Manager`(*注：`cmd/prometheus/main.go`*)入口代码如下：  
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

**说明**：
启动`scrapeManager`: `err := scrapeManager.Run(discoveryManagerScrape.SyncCh())` 
- `discoveryManagerScrape.SyncCh()` 返回`chan map[string][]*targetgroup.Group`。服务发现模块(`discoveryManagerScrape`)会通过此`channel` 向`scrape`模块(`scrapeManager`)发送当前最新的拉取对象的信息(`targetgroup.Group`)
- `scrapeManager.Run(chan map[string][]*targetgroup.Group)`，启动`scrapeManager`，准备接收到当前最新的拉取对象的信息(`targetgroup.Group`)


**Manager.Run方法**  


<br/>  

**Manager.reload-热加载的实际执行者**  
执行过程说明：
- 遍历 `m.targetSets ` 为每个 `job`创建 `scrapePool`实例，并将`scrapePool`实例保存在`m.scrapePools`中。`m.targetSets `暂存的是当前最新的抓取目标，是在  `Manager.Run--> Manager.updateTsets(ts) ` 中进行设置的. 
- 协程并发执行`scrapePool.Sync(groups)`,`scrapePools.Sync(groups)`会将`targetGroup`转换为实际的拉取`metrics`的`target`.
- `m.metrics.targetScrapePools.Inc()`:`Manager.metrics` 是对`scrape`模块的监控指标， `Manager.metrics.targetScrapePools`是`Prometheus`的`scrapePool`实例计数，`counter`类型，`metric names`:`prometheus_target_scrape_pools_total`。创建一个新`scrapePool`实例，则`Manager.metrics.targetScrapePools  + 1` ；同理，创建失败则执行`m.metrics.targetScrapePoolsFailed.Inc()`

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
			// 为每个targetSet创建scrapePool实例
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
  
<br/>  


## 抓取指标

### scrapePool 结构体


主要字段说明： 
| 字段名   | 类型  |说明  |备注    |
| :-----| :---- | :---- | :---- |
| `appendable` | `storage.Appendable` | 存储`storage`| `scrapePool`实例`appendable`被赋值为 `fanoutStorage`<br/> 1. `main`调用`scrape.NewManager`创建`scrape.Manager`实例，`scrape.Manager`的`append`字段被赋值`fanoutStorage`<br/> 2. `scrape.reload()`会将`scrape.Manager`的`append`字段传入`newScrapePool`函数，创建`scrapePool`对象。`scrapePool`实例的`appendable`字段就赋值为`fanoutStorage`|
| `config` | `*config.ScrapeConfig` | 抓取的配置 | |
| `activeTargets` | `map[uint64]*Target`  | 需要抓取的target| |
| `droppedTargets`|`int`  | 需要丢弃的目标,即不需要抓取的target| |
| `droppedTargetsCount`|`[]*Target`  | 不需要抓取的target的数量| |
| `client`|`*http.Client` | `http client`用于`pull`指标时 发起`http`请求|每个`scrapePool`实例只有一个`http client`,向`activeTargets`的多个公用此`http client`|

**思考题**  
1. `activeTargets`、`droppedTargets`都是`Target`集合，为何`activeTargets`以map进行组织，`droppedTargets`选择切片呢？  
todo


### 将targetGroup转换为实际的抓取目标

**方法 scrapePool.Sync**
`Manager.reload()`会为每个拉取的`job`执行`sp.Sync(groups)`(定义:`func (sp *scrapePool) Sync(tgs []*targetgroup.Group)`)。那么`scrapePool.Sync`方法的作用有哪些呢？
- 将`targetgroup.Group`转换为实际的拉取目标
- 整合全部需要抓取目标和丢弃的目标
- 调用`scrapePool.sync`
  - 对需要抓取目标去重;
  - 对于需要抓取的新目标,启动协程进行抓取指标`go l.run(interval, timeout, nil)`;
  - 对于需要丢弃的目标,停止对其抓取工作

```go

// Sync converts target groups into actual scrape targets and synchronizes
// the currently running scraper with the resulting set and returns all scraped and dropped targets.
// 将 targetgroup.Group 转换为实际的拉取目标 
func (sp *scrapePool) Sync(tgs []*targetgroup.Group) {
	sp.mtx.Lock()
	defer sp.mtx.Unlock()
	start := time.Now()

	sp.targetMtx.Lock()
	// 需要拉取的目标
	var all []*Target
	var targets []*Target
	lb := labels.NewBuilder(labels.EmptyLabels())
	// 需要丢弃的目标
	sp.droppedTargets = []*Target{}
	sp.droppedTargetsCount = 0
	for _, tg := range tgs {
		// 将 targetgroup.Group 转换为实际的拉取目标，赋值给变量targets
		targets, failures := TargetsFromGroup(tg, sp.config, sp.noDefaultPort, targets, lb)
		for _, err := range failures {
			level.Error(sp.logger).Log("msg", "Creating target failed", "err", err)
		}
		sp.metrics.targetSyncFailed.WithLabelValues(sp.config.JobName).Add(float64(len(failures)))
		for _, t := range targets {
			// Replicate .Labels().IsEmpty() with a loop here to avoid generating garbage.
			nonEmpty := false
			t.LabelsRange(func(l labels.Label) { nonEmpty = true })  // 检查target是否是需要拉取目标
			switch {
			case nonEmpty:  // 如果是需要拉取目标，暂存于切片all
				all = append(all, t)
			case !t.discoveredLabels.IsEmpty():
				if sp.config.KeepDroppedTargets == 0 || uint(len(sp.droppedTargets)) < sp.config.KeepDroppedTargets {
					sp.droppedTargets = append(sp.droppedTargets, t)  // 如果是丢弃的目标，暂存于切片sp.droppedTargets中，并且计数器 sp.droppedTargetsCount+1
				}
				sp.droppedTargetsCount++
			}
		}
	}
	sp.targetMtx.Unlock()
	sp.sync(all) // 将需要拉取目标进行去重，拉取指标

	sp.metrics.targetSyncIntervalLength.WithLabelValues(sp.config.JobName).Observe(
		time.Since(start).Seconds(),
	)
	sp.metrics.targetScrapePoolSyncsCounter.WithLabelValues(sp.config.JobName).Inc()
}


// sync takes a list of potentially duplicated targets, deduplicates them, starts
// scrape loops for new targets, and stops scrape loops for disappeared targets.
// It returns after all stopped scrape loops terminated.
// 将需要拉取目标进行去重;
// 对于需要抓取的新目标，拉取指标
// 对于需要丢弃的目标,停止对其抓取工作
func (sp *scrapePool) sync(targets []*Target) {
	var (
		uniqueLoops   = make(map[uint64]loop)
		interval      = time.Duration(sp.config.ScrapeInterval)
		timeout       = time.Duration(sp.config.ScrapeTimeout)
		bodySizeLimit = int64(sp.config.BodySizeLimit)
		sampleLimit   = int(sp.config.SampleLimit)
		bucketLimit   = int(sp.config.NativeHistogramBucketLimit)
		labelLimits   = &labelLimits{
			labelLimit:            int(sp.config.LabelLimit),
			labelNameLengthLimit:  int(sp.config.LabelNameLengthLimit),
			labelValueLengthLimit: int(sp.config.LabelValueLengthLimit),
		}
		honorLabels              = sp.config.HonorLabels
		honorTimestamps          = sp.config.HonorTimestamps
		enableCompression        = sp.config.EnableCompression
		trackTimestampsStaleness = sp.config.TrackTimestampsStaleness
		mrc                      = sp.config.MetricRelabelConfigs
		scrapeClassicHistograms  = sp.config.ScrapeClassicHistograms
	)

	sp.targetMtx.Lock()
	// targets就是scrapePool.Sync整理的需要拉取指标的target列表
	for _, t := range targets {
		hash := t.hash()  // 计算每个target的hash值,相同target哈希值相同。使用此hash实现去重

		// activeTargets以map数据结构进行组织，可以实现去重
		if _, ok := sp.activeTargets[hash]; !ok {
			// The scrape interval and timeout labels are set to the config's values initially,
			// so whether changed via relabeling or not, they'll exist and hold the correct values
			// for every target.
			var err error
			interval, timeout, err = t.intervalAndTimeout(interval, timeout)
			s := &targetScraper{
				Target:               t,
				client:               sp.client, // 每个target的拉取 共享scrapePool.client 发起请求
				timeout:              timeout,
				bodySizeLimit:        bodySizeLimit,
				acceptHeader:         acceptHeader(sp.config.ScrapeProtocols),
				acceptEncodingHeader: acceptEncodingHeader(enableCompression),
				metrics:              sp.metrics,
			}
			l := sp.newLoop(scrapeLoopOptions{
				target:                   t,
				scraper:                  s,
				sampleLimit:              sampleLimit,
				bucketLimit:              bucketLimit,
				labelLimits:              labelLimits,
				honorLabels:              honorLabels,
				honorTimestamps:          honorTimestamps,
				enableCompression:        enableCompression,
				trackTimestampsStaleness: trackTimestampsStaleness,
				mrc:                      mrc,
				interval:                 interval,
				timeout:                  timeout,
				scrapeClassicHistograms:  scrapeClassicHistograms,
			})
			if err != nil {
				l.setForcedError(err)
			}

			sp.activeTargets[hash] = t
			sp.loops[hash] = l

			uniqueLoops[hash] = l  // 如果不在activeTargets列表的target，也就是新的target会加入到uniqueLoops；原有的target不会加入到uniqueLoops
		} else {
			// This might be a duplicated target.
			// 因为服务发现组件发送的是当前最新的监控对象的信息，是全量监控对象。
			// 如果存在于activeTargets列表的target,没有在uniqueLoops中，说明是已经存在的target，已经处于监控之中
			// 将已经存在的target存入uniqueLoops 值为nil。
			// 通过值是否为nil即可判断是新加入的target 还是原有的target
			if _, ok := uniqueLoops[hash]; !ok {
				uniqueLoops[hash] = nil
			}
			// Need to keep the most updated labels information
			// for displaying it in the Service Discovery web page.
			sp.activeTargets[hash].SetDiscoveredLabels(t.DiscoveredLabels())
		}
	}

	var wg sync.WaitGroup

	// Stop and remove old targets and scraper loops.
	// 因为sp.activeTargets是前一次更新时全部的target
	// 本次的uniqueLoops是当前最新的全部的target
	// 在sp.activeTargets中,而不在uniqueLoops中的target就视为过时的target,无需进行监控。关闭对其监控，清除其相关信息 
	for hash := range sp.activeTargets {
		if _, ok := uniqueLoops[hash]; !ok {
			wg.Add(1)
			go func(l loop) {
				l.stop()
				wg.Done()
			}(sp.loops[hash])

			delete(sp.loops, hash)
			delete(sp.activeTargets, hash)
		}
	}

	sp.targetMtx.Unlock()

	sp.metrics.targetScrapePoolTargetsAdded.WithLabelValues(sp.config.JobName).Set(float64(len(uniqueLoops)))
	forcedErr := sp.refreshTargetLimitErr()
	for _, l := range sp.loops {
		l.setForcedError(forcedErr)
	}

	// uniqueLoops是当前需要拉取的target，而且已经去重过了
	// 在uniqueLoops获取到的值为nil，说明是之前已经存在的target，已经处于监控之中，本次无需拉取指标
	for _, l := range uniqueLoops {  
		if l != nil {
			go l.run(nil) // 为每个target调用 scrapeLoop.run方法,拉取指标。scrapeLoop.run  --> scrapeLoop.scrapeAndReport
		}
	}
	// Wait for all potentially stopped scrapers to terminate.
	// This covers the case of flapping targets. If the server is under high load, a new scraper
	// may be active and tries to insert. The old scraper that didn't terminate yet could still
	// be inserting a previous sample set.
	wg.Wait()
}
```

**拉取操作** 

**注： 阅读此部分代码，前提条件是了解`prometheus`的[`storage`模块](./005.storage.md)**  


调用链：`scrapeLoop.sync` --> `scrapeLoop.run` --> `scrapeLoop.scrapeAndReport`

`scrapeLoop.scrapeAndReport`说明：
- 拉取`target`的指标，即:`resp, scrapeErr = sl.scraper.scrape(scrapeCtx) `
- 写`scrape`的数据写入底层存储,即:`total, added, seriesAdded, appErr = sl.append(app, b, contentType, appendTime)`

todo：与存储共同阐述 
代码解析：
```go

func (sl *scrapeLoop) run(errc chan<- error) {
	if !sl.skipOffsetting {
		select {
		case <-time.After(sl.scraper.offset(sl.interval, sl.offsetSeed)):
			// Continue after a scraping offset.
		case <-sl.ctx.Done():
			close(sl.stopped)
			return
		}
	}

	var last time.Time

	alignedScrapeTime := time.Now().Round(0)
	ticker := time.NewTicker(sl.interval)
	defer ticker.Stop()

mainLoop:
	for {
		select {
		case <-sl.parentCtx.Done():
			close(sl.stopped)
			return
		case <-sl.ctx.Done():
			break mainLoop
		default:
		}

		// Temporary workaround for a jitter in go timers that causes disk space
		// increase in TSDB.
		// See https://github.com/prometheus/prometheus/issues/7846
		// Calling Round ensures the time used is the wall clock, as otherwise .Sub
		// and .Add on time.Time behave differently (see time package docs).
		scrapeTime := time.Now().Round(0)
		if AlignScrapeTimestamps && sl.interval > 100*ScrapeTimestampTolerance {
			// For some reason, a tick might have been skipped, in which case we
			// would call alignedScrapeTime.Add(interval) multiple times.
			for scrapeTime.Sub(alignedScrapeTime) >= sl.interval {
				alignedScrapeTime = alignedScrapeTime.Add(sl.interval)
			}
			// Align the scrape time if we are in the tolerance boundaries.
			if scrapeTime.Sub(alignedScrapeTime) <= ScrapeTimestampTolerance {
				scrapeTime = alignedScrapeTime
			}
		}

		last = sl.scrapeAndReport(last, scrapeTime, errc) // 抓取

		select {
		case <-sl.parentCtx.Done():
			close(sl.stopped)
			return
		case <-sl.ctx.Done():
			break mainLoop
		case <-ticker.C:
		}
	}

	close(sl.stopped)

	if !sl.disabledEndOfRunStalenessMarkers {
		sl.endOfRunStaleness(last, ticker, sl.interval)
	}
}




// scrapeAndReport performs a scrape and then appends the result to the storage
// together with reporting metrics, by using as few appenders as possible.
// In the happy scenario, a single appender is used.
// This function uses sl.appenderCtx instead of sl.ctx on purpose. A scrape should
// only be cancelled on shutdown, not on reloads.
func (sl *scrapeLoop) scrapeAndReport(last, appendTime time.Time, errc chan<- error) time.Time {
	start := time.Now()

	// Only record after the first scrape.
	if !last.IsZero() {
		sl.metrics.targetIntervalLength.WithLabelValues(sl.interval.String()).Observe(
			time.Since(last).Seconds(),
		)
	}

	var total, added, seriesAdded, bytesRead int
	var err, appErr, scrapeErr error

	app := sl.appender(sl.appenderCtx)
	defer func() {
		if err != nil {
			app.Rollback()
			return
		}
		err = app.Commit()
		if err != nil {
			level.Error(sl.l).Log("msg", "Scrape commit failed", "err", err)
		}
	}()

	defer func() {
		if err = sl.report(app, appendTime, time.Since(start), total, added, seriesAdded, bytesRead, scrapeErr); err != nil {
			level.Warn(sl.l).Log("msg", "Appending scrape report failed", "err", err)
		}
	}()

	if forcedErr := sl.getForcedError(); forcedErr != nil {
		scrapeErr = forcedErr
		// Add stale markers.
		if _, _, _, err := sl.append(app, []byte{}, "", appendTime); err != nil {
			app.Rollback()
			app = sl.appender(sl.appenderCtx)
			level.Warn(sl.l).Log("msg", "Append failed", "err", err)
		}
		if errc != nil {
			errc <- forcedErr
		}

		return start
	}

	var contentType string
	var resp *http.Response
	var b []byte
	var buf *bytes.Buffer
	scrapeCtx, cancel := context.WithTimeout(sl.parentCtx, sl.timeout)

	resp, scrapeErr = sl.scraper.scrape(scrapeCtx)  // 拉取target的指标
	if scrapeErr == nil {
		b = sl.buffers.Get(sl.lastScrapeSize).([]byte)
		defer sl.buffers.Put(b)
		buf = bytes.NewBuffer(b)
		contentType, scrapeErr = sl.scraper.readResponse(scrapeCtx, resp, buf)
	}
	cancel()

	if scrapeErr == nil {
		b = buf.Bytes()
		// NOTE: There were issues with misbehaving clients in the past
		// that occasionally returned empty results. We don't want those
		// to falsely reset our buffer size.
		if len(b) > 0 {
			sl.lastScrapeSize = len(b)
		}
		bytesRead = len(b)
	} else {
		level.Debug(sl.l).Log("msg", "Scrape failed", "err", scrapeErr)
		if errc != nil {
			errc <- scrapeErr
		}
		if errors.Is(scrapeErr, errBodySizeLimit) {
			bytesRead = -1
		}
	}

	// A failed scrape is the same as an empty scrape,
	// we still call sl.append to trigger stale markers.
	total, added, seriesAdded, appErr = sl.append(app, b, contentType, appendTime)
	if appErr != nil {
		app.Rollback()
		app = sl.appender(sl.appenderCtx)
		level.Debug(sl.l).Log("msg", "Append failed", "err", appErr)
		// The append failed, probably due to a parse error or sample limit.
		// Call sl.append again with an empty scrape to trigger stale markers.
		if _, _, _, err := sl.append(app, []byte{}, "", appendTime); err != nil {
			app.Rollback()
			app = sl.appender(sl.appenderCtx)
			level.Warn(sl.l).Log("msg", "Append failed", "err", err)
		}
	}

	if scrapeErr == nil {
		scrapeErr = appErr
	}

	return start
}



```
 -->
