# fasthttp为什么那么快?

`fasthttp`并不是万金油。

## HTTP框架的原理

我们知道，`HTTP`协议是建立在`TCP`协议基础上的，所以`HTTP`框架本质就是处理`TCP`链接和`TCP`的`payload`。

<img src='./src/HTTP框架的原理.drawio.png'>

`http server`工作流程：

1. 创建`socket`、绑定端口并监听、建立`TCP`链接和获取数据.
2. 按照`http`协议的格式解析获取到的数据，解析失败则反汇编
