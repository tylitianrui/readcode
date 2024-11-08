# 2.5 二开：prometheus 热更新



##  1.`prometheus` 官方提供的热更新方式

`prometheus` 官方提供了两种热更新的方法：

方法一： 向`prometheus`进程发送系统信号`HUP` ，`prometheus`即可实现热更新。

``` shell
kill   -HUP  18956
```

`Prometheus PID` 需要用户自行查找。例如使用`ps` 命令查找`Prometheus`进程号

 ``````shell
 ps aux | grep prometheus
 ``````



方法二：通过调用`api`接口实现热更新

如果使用此方法进行热更新。`prometheus`启动时，必须添加参数 `--web.enable-lifecycle`

``````
./prometheus  --config.file=documentation/examples/prometheus.yml   --web.enable-lifecycle
``````

调用`prometheus`的api接口`/-/reload`，`prometheus`即可实现热更新。

``````shell
curl -X POST http://127.0.0.1:9090/-/reload 
# 或者使用 put 方法
curl -X PUT http://127.0.0.1:9090/-/reload 

curl -X PUT http://127.0.0.1:9090/-/quit 
``````



除了`prometheus`重启的`api`接口`/-/reload`之外，`prometheus`还提供了关闭`prometheus`进程`api`接口`/-/quit`接口。这些接口和第三方查询监控指标的接口是没有隔离的，不安全。



## 2. 二开：安全的热更新

###  需求

- 执行`prometheus relaod` 命令进行热更新

- 执行`prometheus stop` 命令关闭`prometheus`进程

- 兼容现有`prometheus`所有命令

- `/-/reload` 接口`/-/quit`接口 禁止调用

  

  

### 方案

`prometheus` 官方版本监听到`HUP`，则进行热更新； 监听到`TERM`信号，则进行关闭。即：

- 热更新

  ``` shell
  kill   -HUP  <Prometheus PID>
  ```

- 关闭`prometheus`进程

  ``````shell
  kill -TERM   <Prometheus PID>
  ``````

  

我对`prometheus` 进行二开：

- `prometheus` 启动时，创建`prometheus.pid`文件 并且记录`prometheus` 进程号；`prometheus` 退出时，删除`prometheus.pid`
- 执行`prometheus relaod` ，则读取`prometheus.pid`文件进程号，并且对此进程发送`HUP`信号
- 执行`prometheus stop` ，则读取`prometheus.pid`文件进程号，并且对此进程发送`TERM`信号
- 调用`/-/reload` 接口`/-/quit`接口则 返回403 和 禁用信息。

 
### 开发

代码分支[develop/hot_relaod](https://github.com/tylitianrui/prometheus/tree/develop/hot_relaod)