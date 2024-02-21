# 1. 项目结构

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
  │      ├── manager.go    
  │      ├── registry.go
  │      ....
  ├── model   
  ├── notifier              notifier 告警模块
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

# 2. main函数分析

## 2.1 goroutine管理(第三方依赖)

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



## 2.2 main函数执行

### 2.2.1 执行流程图  
<br/> 

![main函数执行](./src/prometheus-main-执行.drawio.png)


说明：  
- `Agent`模式:  目前最流行的全局视图解决方案就是远程写入`Remote Write`。`Agent`模式优化了远程写入方案。`Prometheus` 在`Agent`模式下,**查询**、**告警**和**本地存储**是被禁用的，开启`TSDB WAL`功能；其他功能保持不变:**抓取逻辑**、**服务发现**等。 详见[**Agent模式**](./代理模式.md)
- `Scrape discovery manager` 与 `Scrape manager` 关系： `Scrape discovery manager`负责服务发现，会不断获取当前最新的服务地址、数量等信息；`Scrape manager`会根据`Scrape discovery manager`获取的节点信息，更新需要采集的`targets`，实现采集。`Notify discovery manager`与` Notifier`亦同
