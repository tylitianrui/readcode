# fasthttp为什么那么快?

`fasthttp`是一个高性能的`web server`框架。`fasthttp`的性能可以达到标准库`net/http`的`10`倍以上。

`fasthttp`并不是万金油，它适用于一些**高并发**场景。如果不考虑**高性能**，`golang`标准库`net/http`才是更优的选择。因为`golang`标准库`net/http` 使用简单易用，便于开发。

`fasthttp`和标准库`net/http`的主要区别：

- `fasthttp` 不支持请求的路由功能。如果用路由功能，要么使用第三方依赖，要么自己实现。对于

  ``````go
  // net/http code
  
  m := &http.ServeMux{}
  m.HandleFunc("/foo/name", fooHandlerFunc)
  m.HandleFunc("/bar", barHandlerFunc)
  m.Handle("/baz", bazHandler)
  
  http.ListenAndServe(":80", m)
  
  ``````

  ``````go
  // the corresponding fasthttp code
  handler := func(ctx *fasthttp.RequestCtx) {
  	switch string(ctx.Path()) {
  	case "/foo/name":
  		fooHandlerFunc(ctx)
  	case "/bar":
  		barHandlerFunc(ctx)
  	case "/baz":
  		bazHandler.HandlerFunc(ctx)
  	default:
  		ctx.Error("not found", fasthttp.StatusNotFound)
  	}
  }
  
  fasthttp.ListenAndServe(":80", handler)
  ``````

- `fasthttp`  与标准库`net/http`的`API` 是不相同的，所以两者不能平替。

- `fasthttp` 用法更加严格。例如`fasthttp` 不允许请求处理函数中持有对`RequestCtx`的引用,会导致内存泄露或其他不可预期的行为。因为`fasthttp`是复用`RequestCtx`对象。如果需要在请求处理函数中使用`RequestCtx`的数据。需要把相关的数据复制出来到其他的数据结构中，避免持有`RequestCtx`对象的引用

- `HTTP`是一个文本格式的协议。`fasthttp` 更倾向于使用`[]byte`  而不是`string`。所以在研习代码和编写代码时会增加一定困难。

  

> [!NOTE]
>
> `fasthttp` 是和标准库`net/http`对标的。工作中，常用的`web` 框架，例如`Gin`、  `Fiber` 分别是基于`net/http`和`fasthttp`实现的。




## HTTP框架的原理 

我们知道，`HTTP`协议是建立在`TCP`协议基础上的，所以`HTTP`框架本质就是处理`TCP`链接和`TCP`的`payload`。

<img src='./src/HTTP框架的原理.drawio.png'>

如果小伙伴们从`0`开始开发一个`http server`框架，必须处理以下开发工作：

1. 创建`socket`、绑定端口并监听、循环获取请求的字节流;
2. 按照`http`协议的格式解析获取到的数据，解析失败，返回用户错误信息；
3. 根据`Method`和`URL`将请求路由到处理函数;
4. 处理请求,并响应。



`fasthttp` 参与的阶段：

- 创建`socket`、绑定端口并监听、循环获取请求的字节流;
- 按照`http`协议的格式解析获取到的数据。
- 处理`http`请求和响应


## 高效的底层设计

`fasthttp` 和标准库`net/http`底层设计不同。

### 标准库`net/http`底层原理

TODO 

### `fasthttp`底层原理

TODO 


## 复用




## 数据类型的选择
