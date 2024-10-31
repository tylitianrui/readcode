# 2.4 prometheus启动流程-main函数分析

在分析之前，我们需要先解决两个问题：

1. `prometheus`代码目录结构
2. 各模块goroutine的编排 管理。有赖于第三方依赖 [github.com/oklog/run](https://github.com/oklog/run)

## 1. 代码的目录结构

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


## 2. 各模块goroutine的编排管理

在`prometheus`中，使用了很多第三方库，为什么要单独说明这个依赖呢？ 因为`prometheus`所有组件`goroutine`都是通过依赖`github.com/oklog/run`(*注：下文简称`run`,代码仓库[https://github.com/oklog/run](https://github.com/oklog/run)*) 进行编排管理的。

### **API说明**

`run`有一个类型`Group`,`Group`有两个公共方法：`Add`、`Run`。

#### `func (g *Group) Add(execute func() error, interrupt func(error))` 

将函数`execute`、`interrupt`注到`Group`对象，注册阶段并不会执行`execute`、`interrupt`函数。函数`execute`、`interrupt`是需要开发伙自己去开发的。

- 函数`execute`：实际需要执行的业务逻辑。
- 函数`interrupt`：执行资源回收、关闭网络连接、关闭文件句柄等**收尾**工作。

#### `func (g *Group) Run() error`   

为每个执行函数`execute`都开启一个独立的`goroutine`去运行。如果有某一个执行函数`execute`报错，所有的退出函数`interrupt`都会接受到这个错误。程序执行资源回收、关闭网络连接、关闭文件句柄等**收尾**工作



**demo**

[代码](https://github.com/tylitianrui/readcode/tree/master/prometheus/run_demo)

``````go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
)

func main() {
	var g run.Group
	term := make(chan os.Signal, 1)
	cancel := make(chan struct{})
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	time1 := NewXtimer("Xtimer1")
	time2 := NewXtimer("Xtimer2")

	g.Add(
		func() error {
			select {
			case sig := <-term:
				fmt.Println("接收到系统信号", sig.String())
			case <-cancel:
				fmt.Println("cancel 有信号了")
			}
			return fmt.Errorf("接收到系统信号")
		},
		func(err error) {
			fmt.Println("信号监听关闭")
			close(cancel)
		},
	)

	g.Add(
		time1.PrintTime, time1.Stop,
	)
	g.Add(
		time2.PrintTime, time2.Stop,
	)
	if err := g.Run(); err != nil {
		os.Exit(1)
	}

}

type Xtimer struct {
	Name   string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewXtimer(name string) *Xtimer {
	ctx, cancel := context.WithCancel(context.TODO())
	return &Xtimer{
		Name:   name,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *Xtimer) PrintTime() error {
	for {
		select {
		case <-t.ctx.Done():
			fmt.Println(t.Name, "退出")
			// return nil
			return fmt.Errorf("%v stop", t.Name)
		default:
			time.Sleep(2 * time.Second)
			fmt.Println(t.Name, time.Now())
		}
	}

}

func (t *Xtimer) Stop(err error) {
	t.cancel()
}

``````

执行结果

``````text
Xtimer1 2024-11-01 00:42:47.592574 +0800 CST m=+6.004138239
Xtimer2 2024-11-01 00:42:47.592623 +0800 CST m=+6.004186926
Xtimer2 2024-11-01 00:42:49.593881 +0800 CST m=+8.005391548
Xtimer1 2024-11-01 00:42:49.593924 +0800 CST m=+8.005435090
Xtimer1 2024-11-01 00:42:51.59523 +0800 CST m=+10.006687255
Xtimer2 2024-11-01 00:42:51.595305 +0800 CST m=+10.006761822
^C
接收到系统信号 interrupt
信号监听关闭
Xtimer1 2024-11-01 00:42:53.596289 +0800 CST m=+12.007691938
Xtimer2 2024-11-01 00:42:53.59632 +0800 CST m=+12.007723791
Xtimer2 退出
Xtimer1 退出
exit status 1
``````



### **代码执行原理**

示意图如下

<img src="./src/run执行流程.drawio.png" width="100%" height="100%" alt="offset默认">

1. 函数`execute`、函数`interrupt` 都是成对出现的，这一对函数被称为一个`actor`。函数`execute`、函数`interrupt` 是开发者编写的。**编码的时候要求：如果退出函数`interrupt`被调用，那么对应的`execute`函数必须能感知到，并且退出**
2. `Add`方法会把这一对`actor`函数注册到`Group`类型对象里。`Group`类型对象里维护一个`actor`类型的切片(*注：切片go语言里的可变长数组，非go语言开发者理解为数组即可*)
3. `Run`方法
   1. 创建一个容量`chan error` ,容量与`actor`切片长度相等。
   2. 为每个函数`execute`都开启一个独立的`goroutine`去运行。如果某一个函数`execute`执行报错，错误就会被发送到`chan error` 
   3. 监听`chan error`。如果没有监听到错误，程序被阻塞；如果监听到错误，程序退出阻塞状态，执行后续的退出逻辑。
   4.  如果接收的error 是从`chan error`里接收的第一个`error`,遍历执行所有的`interrupt` 函数；
   5. `interrupt` 函数执行，对应的函数`execute`就会退出。函数`execute`退出时会向`chan error` 发生一个`error`类型的数据。退出逻辑接收到所有函数`execute`退出时的error，则认为函数`execute`全部退出。
   6. 程序关闭

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




## 3. main函数执行流程分析

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