# fastHTTP为什么那么快？

## HTTP框架的原理

我们知道，`HTTP`协议是建立在`TCP`协议基础上的，所以`HTTP`框架本质就是处理`TCP`链接和`TCP`的`payload`。

<img src='./src/HTTP框架的原理.drawio.png'>

对应服务端，通过`TCP`获取数据之后，需要以
