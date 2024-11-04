# 2.5 二开：prometheus 热更新



##  1.`prometheus` 官方提供的热更新方式

`prometheus` 官方提供了两种热更新的方法：

方法一： 向`prometheus`进程发送系统信号`HUP` ，`prometheus`即可实现热更新。

``` shell
kill   -HUP  <Prometheus PID>
```

`Prometheus PID` 需要用户自行查找。例如使用`ps` 命令查找`Prometheus`进程号

 ``````shell
 ps aux | grep prometheus
 ``````



方法二：通过调用`api`接口实现热更新

如果使用此方法进行热更新。`prometheus`启动时，必须添加参数 `--web.enable-lifecycle`

``````
./prometheus  --config.file=/Users/ollie/opencode/prometheus/documentation/examples/prometheus.yml   --web.enable-lifecycle
``````

调用`prometheus`的api接口`/-/reload`，`prometheus`即可实现热更新。

``````shell
curl -X POST http://127.0.0.1:9090/-/reload 
# 或者使用 put 方法
curl -X PUT http://127.0.0.1:9090/-/reload 
``````



除了`prometheus`重启的`api`接口`/-/reload`之外，`prometheus`还提供了关闭`prometheus`进程`api`接口`/-/quit`接口。这些接口和第三方查询监控指标的接口是没有隔离的，不安全。



## 2. 二开：安全的热更新



















