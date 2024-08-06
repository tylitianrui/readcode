# prometheus server

## 1. 项目结构

项目代码结构: 
```
prometheus
  ├── cmd                  程序入口
  │   ├── prometheus
  │   │    └── main.go     prometheus入口
  │   └── promtool
  │        └── main.go     promtool 入口
  ├── config               解析配置文件
  ├── discovery            服务发现模块
  │      ├── aws           aws服务发现模块
  │      ├── kubernetes    kubernetes服务发现模块
  │      ...   
  │      ├── registry.go
  │      ├── discoverer_metrics_noop.go
  │      ├── discovery.go   
  │      ├── manager.go
  │      ├── metrics.go
  │      ├── metrics_k8s_client.go
  │      ├── metrics_refresh.go   
  │      └── util.go   
  ├── model   
  ├── notifier            notifier 告警模块
  ├── plugins
  ├── prompb
  ├── promql               promql查询实现
  ├── rules                规则管理模块
  ├── scrape               拉取指标等相关
  ├── storage              存储相关 remote storage 和 local storage
  ├── tracing   
  ├── tsdb                 tsdb数据库
  └── util    

```

## 3.prometheus server启动-main函数分析

### 3.1 goroutine管理(第三方依赖)

代码仓库: [github.com/oklog/run](https://github.com/oklog/run)  

在`prometheus`中，使用了很多第三方库，为什么要单独说明这个依赖呢？ 因为`prometheus`所有组件`goroutine`都是通过此依赖进行管理的。  

下面是`main`函数中一段代码: 
```go
	var g run.Group
  // ....
  	{
		// Scrape discovery manager.
		g.Add(
			func() error {
				err := discoveryManagerScrape.Run()
				level.Info(logger).Log("msg", "Scrape discovery manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping scrape discovery manager...")
				cancelScrape()
			},
		)
	}
  // ....
  if err := g.Run(); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
```

<br/>

`github.com/oklog/run` 有两个主要方法:
- `func (g *Group) Add(execute func() error, interrupt func(error))`  注册执行函数`execute`、退出函数`interrupt`，为每个执行函数`execute` 都开启一个独立的`goroutine`去运行。如果有某一个执行函数`execute`报错，就会执行所有的退出函数`interrupt`进行退出、回收等“善后”工作。
- `func (g *Group) Run() error ` 运行



### 3.2 main函数执行

#### 3.2.1 执行流程图  

<br/> 

`prometheus` 启动流程  

![main函数执行](./src/prometheus-main-执行.drawio.png)


`prometheus` 流程  



说明：  
- `Agent`模式:  目前最流行的全局视图解决方案就是远程写入`Remote Write`。`Agent`模式优化了远程写入方案。`prometheus` 在`Agent`模式下,**查询**、**告警(`Rule manager`、`Notifier`)**和**本地存储(`Rule manager`、`TSDB`)**是被禁用的，开启`TSDB WAL`功能；其他功能保持不变:**抓取逻辑**、**服务发现**等。 详见[**Agent模式**](./代理模式.md)
- `Scrape discovery manager`负责服务发现，会不断获取当前最新的服务地址、数量等信息；
- `Scrape manager`组件会通过`Scrape discovery manager`组件获取的服务节点信息，更新需要采集的`targets`，进行采集指标
- `Rule manager`:`prometheus`进行规则管理的组件。分为两类规则：
  - `RecordingRule` 表达式规则，用于记录到tsdb;
  - `AlertingRule`  告警规则,触发告警
- `Notifier`：用于向 `Alertmanager` 发送告警通知
- `Termination handler`: 监听信号`os.Interrupt`(注：`ctrl+c`，`kill -2 p<prometheus pid>` )、`syscall.SIGTERM`(注：`kill -15  <prometheus pid>`),则优雅地退出`prometheus`进程。
- `Reload handler`: 监听信号`syscall.SIGHUP`(注：`kill -HUP  <prometheus pid>`，`kill -1  <prometheus pid>` ),则重新加载配置文件。