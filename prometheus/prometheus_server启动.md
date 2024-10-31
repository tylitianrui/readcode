# 2.4 prometheus启动流程-main函数分析

在分析之前，我们需要先解决两个问题：

1. `prometheus`代码目录结构
2. 各模块goroutine的编排 管理。有赖于第三方依赖 [github.com/oklog/run](https://github.com/oklog/run)

## 代码的目录结构

项目代码结构  

```text
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
  │      ├── 
  │      ├── discoverer_metrics_noop.go
  │      ├── discovery.go   
  │      ├── manager.go
  │      ├── metrics.go
  │      ├── metrics_k8s_client.go
  │      ├── metrics_refresh.go   
  │      └── util.go   
  ├── model   
  ├── notifier       notifier 告警模块
  ├── plugins        插件
  ├── prompb         proto文件
  ├── promql         promql查询实现
  ├── rules          规则管理模块。告警规则、记录规则的实现
  ├── scrape         拉取指标等相关代码
  ├── storage        存储相关代码。存储代理层
  ├── tracing   
  ├── tsdb           tsdb数据库
  ├── web                  
  │     ├── api      Prometheus API实现，http server   
  │     └── ui       Prometheus UI界面，前端
  └── util           工具类

```

在[项目简述与准备](./项目简述与准备.md)部分，代码编译时会创建两个二进制文件：`prometheus`、`promtool`，这两二进制文件入口函数分别是`cmd/prometheus/main.go`、`cmd/promtool/main.go`。本节的重点就是解析`cmd/prometheus/main.go`执行过程。


## 各模块goroutine的管理(第三方依赖`github.com/oklog/run`)

在`prometheus`中，使用了很多第三方库，为什么要单独说明这个依赖呢？ 因为`prometheus`所有组件`goroutine`都是通过此依赖(*[代码仓库](https://github.com/oklog/run)*) 进行编排管理的。

**代码仓库**

 https://github.com/oklog/run



**原理**

 [github.com/oklog/run](https://github.com/oklog/run)执行示意图如下

<img src="./src/run执行流程.drawio.png" width="100%" height="100%" alt="offset默认">



**API说明**

- `func (g *Group) Add(execute func() error, interrupt func(error))`

  - 执行函数`execute`：实际需要执行的工作
  - 退出函数`interrupt`：退出函数`interrupt`进行退出、回收等“善后”工作

- `func (g *Group) Run() error` 

   为每个执行函数`execute`都开启一个独立的`goroutine`去运行。如果有某一个执行函数`execute`报错，所有的退出函数`interrupt`都会接受到这个错误。程序可根据错误处理退出、回收资源等“善后”工作。



**补充**

- [Release Party | Ways To Do Things with Peter Bourgon](https://www.youtube.com/watch?v=LHe1Cb_Ud_M&t=1376s)

  

**在prometheus中的应用**


下面是`prometheus main`函数中一段代码:

```golang
 import "github.com/oklog/run"

    
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

   {
		// Notify discovery manager.
		g.Add(
			func() error {
				err := discoveryManagerNotify.Run()
				level.Info(logger).Log("msg", "Notify discovery manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping notify discovery manager...")
				cancelNotify()
			},
		)
	 }
  // ....
  if err := g.Run(); err != nil {
    level.Error(logger).Log("err", err)
    os.Exit(1)
  }
```




## main函数执行

### 执行流程图  

`prometheus` 启动流程如图

![main函数执行](./src/prometheus-main-执行.drawio.png)


`prometheus` 流程说明： 

- `Agent`模式:  目前最流行的全局视图解决方案就是远程写入`Remote Write`。`Agent`模式优化了远程写入方案。`prometheus` 在`Agent`模式下,**查询**、**告警(`Rule manager`、`Notifier`)**和**本地存储(`Rule manager`、`TSDB`)**是被禁用的，开启`TSDB WAL`功能；其他功能保持不变:**抓取逻辑**、**服务发现**等。 详见[**Agent模式**](./代理模式.md)
- `Scrape discovery manager`负责服务发现，会不断获取当前最新的服务地址、数量等信息；
- `Scrape manager`组件会通过`Scrape discovery manager`组件获取的服务节点信息，更新需要采集的`targets`，进行采集指标
- `Rule manager`:`prometheus`进行规则管理的组件。分为两类规则：
  - `RecordingRule` 表达式规则，用于记录到tsdb;
  - `AlertingRule`  告警规则,触发告警
- `Notifier`：用于向 `Alertmanager` 发送告警通知
- `Termination handler`: 监听信号`os.Interrupt`(注：`ctrl+c`，`kill -2 p<prometheus pid>` )、`syscall.SIGTERM`(注：`kill -15  <prometheus pid>`),则优雅地退出`prometheus`进程。
- `Reload handler`: 监听信号`syscall.SIGHUP`(注：`kill -HUP  <prometheus pid>`，`kill -1  <prometheus pid>` ),则重新加载配置文件。