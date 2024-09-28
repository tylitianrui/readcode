# Label和Relabeling

## 标签Label

### 内置标签/Meta标签

一般情况，`prometheus`以 "`__`"作为前缀的标签的是系统内置标签,定义了`Prometheus`的`Target`实例和指标的一些基本信息。
常见内部标签如下：

- `__address__`   当前`Target`实例的访问地址<`host`>:<`port`>，只供`Prometheus`使用，不会写入时序数据库中，也无法使用`promql`查询。
- `__scheme__`    采集`Target`指标的协议，`HTTP`或者`HTTPS` 默认是`HTTP`，只供`Prometheus`使用，不会写入时序数据库中，也无法使用`promql`查询。
- `__metrics_path__`  `Target`对外暴露的采集接口，默认`/metrics`，只供`Prometheus`使用，不会写入时序数据库中，也无法使用`promql`查询。
- `__name__`      `metrics`的名称，指标名会以 `__name__= <metric_name>`的形式，存储在时序数据库中，例如`__name__ = prometheus_http_requests_total`
- `job`   指标归属哪个 `job`
- `instance`   采集的实例，默认情况与`__address__`相同



> [!NOTE]
>
> 内置标签没有定义在`prometheus`项目，定义在`prometheus` 的公共依赖项目 [common](https://github.com/prometheus/common/blob/main/model/labels.go#L42)



### 自定义标签

`prometheus` 允许用户根据自己需求去定义标签，实现多维度、精细查询。使用关键字`labels`创建自定义标签。

#### 案例1:基本使用

要求： 针对当前`prometheus`得监控，为其指标标注归属部门`infra`、服务类型`monitor`、运行环境`prod`等  

配置如下： 

```yaml
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
        labels:
           env: prod
           service: monitor
           biz: infra  
```

展示效果：  

首先，我们先看一下`target`上有哪些`label`,访问[http://127.0.0.1:9090/targets?search=](http://127.0.0.1:9090/targets?search=)  如图：  

![prometheus_label_demo_1_target](./src/prometheus_label_demo_1_target.png)

<br>

我们再看一下指标，**可任选指标**，本次选取`go_memstats_heap_alloc_bytes`展示  

![prometheus_label_demo_1](./src/prometheus_label_demo_1.png) 可见所有指标都被打上这些`env: prod`、`service: monitor`、`biz: infra`标签  



#### 案例2: 作用范围是target，而不是job

在[案例1:基本使用](#案例1基本使用)的基础上，再追加一个`target`目标为`http://127.0.0.1:9090/metrics`,`biz`标签设置为`internal`。两个相同的`target`,标签不同会是什么结果呢？

配置  

```yaml
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
        labels:
           env: prod
           service: monitor
           biz: infra  
      - targets: ["127.0.0.1:9090"]
        labels:
           env: prod
           service: monitor
           biz: internal
```

展示效果：
<br>
首先，我们先看一下`target`上有哪些`label`,访问[http://127.0.0.1:9090/targets?search=](http://127.0.0.1:9090/targets?search=), 如图:  

![prometheus_label_demo_2_target](./src/prometheus_label_demo_2_target.png)



我们再看一下指标，**可任选指标**，本次选取`go_memstats_heap_alloc_bytes`展示 , 如图:  
![prometheus_label_demo_2](./src/prometheus_label_demo_2.png)

可见：两个target获取的指标被打上不同的标签。**`labels`只作用于当前的`target`**

## Relabeling

`Relabeling`是`prometheus`的一种强大的功能，可以在拉取`targets`指标之前，动态地重写、增加、删除标签。 每个`scrape_config`中可以配置多个标签。它们会按照在配置文件中出现的先后顺序而作用与每个目标的标签集。

### 基本使用

`Relabeling` 需要配置在`prometheus`的配置文件中(*例如：`prometheus.yaml`*)。与服务发现配置为同一层级的`relabel_configs`模块下进行配置。
配置的关键字：

- `source_labels` 源标签，没有经过`relabel`处理之前的标签名字。
- `target_label` 目标标签，通过`relabel`处理之后的标签名字。
- `separator` 源标签的值的连接分隔符。默认是"`;`"。
- `regex` 正则表达式，匹配源标签的值默认是`(.*)`。
- `replacement`通过分组替换后标签（`target_label`）对应的值。默认是`$1`
- 具体处理的行为,即`action`，即如果`source_labels`指标满足`regex`规则，那么`prometheus`会进行“特定的处理”，将处理结果赋值给`target_label`。具体有哪些行为呢？ 如下表所示：  

<br>

| action    | 说明                                                         |
| :-------- | :----------------------------------------------------------- |
| replace   | 根据`regex`来去匹配`source_labels`标签上的值，并将改写到`target_label`中标签。如果未指定`action`，则默认就是`replace` |
| keep      | 根据`regex`来去匹配`source_labels`标签上的值，如果匹配成功，则采集此`target`,否则不采集 |
| drop      | 根据`regex`来去匹配`source_labels`标签上的值，如果匹配成功，则不采集此`target`,用于排除，与`keep`相反 |
| labelkeep | 使用`regex`表达式匹配标签，仅收集符合规则的`target`，不符合匹配规则的不收集 |
| labeldrop | 使用`regex`表达式匹配标签，符合规则的标签将从`target`实例中移除 |
| labelmap  | 根据`regex`的定义去匹配`Target`实例所有标签的名称，并且以匹配到的内容为新的标签名称，其值作为新标签的值 |



#### Relabeling - replace 标签替换

####  案例3: replace基本使用

在[案例2: 作用范围是target，而不是job](#案例2: 作用范围是target，而不是job)的基础上，将标签`service` 的值改下成`prometheus_monitor`。配置如下 

``````yaml
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
        labels:
           env: prod
           service: monitor
           biz: infra  
      - targets: ["127.0.0.1:9090"]
        labels:
           env: prod
           service: monitor
           biz: internal
    relabel_configs:
    - source_labels:
      - "service"
      target_label: "service"
      action: replace
      replacement: prometheus_monitor
``````



说明： 要处理的源标签(*配置`source_labels`*)`service`。如果标签`service`的值匹配正则`(.*)`,那么将配置`replacement`的值(*此例子中为常量`prometheus_monitor`*) 赋值给目标标签(*配置`target_label`*) ` service`。

展示

![replace基本使用](./src/relabel_replace_from_service_to_service_1.png)



我们再看一下指标，**可任选指标**，本次选取`go_memstats_heap_alloc_bytes`展示 , 如图:  

![replace基本使用](./src/relabel_replace_from_service_to_service_2.png)



#### 案例4: 使用replace新增自定义标签

在[案例2: 作用范围是target，而不是job](#案例2: 作用范围是target，而不是job)的基础上，将内置标签`__address__` 的`ip`地址部分改写成自定义标签`node`,并存入数据库。配置如下 

``````yaml
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
        labels:
           env: prod
           service: monitor
           biz: infra  
      - targets: ["127.0.0.1:9090"]
        labels:
           env: prod
           service: monitor
           biz: internal
    relabel_configs:
    - source_labels:
      - "__address__"
      regex: "(.*):(.*)"
      target_label: "node"
      action: replace
      replacement: $1
``````



说明：

- 内置标签 `__address__`  只供`Prometheus`内部使用，不会写入时序数据库中，也无法使用`promql`查询。
- 如果源标签 `__address__`的值匹配正则匹配`(.*):(.*)`,那么将`$1`位置上的值(*即:`ip`*部分) 赋值给自定义标签`node`；如果源标签 `__address__`的值不能匹配正则匹配`(.*):(.*)`,不进行赋值



展示

![使用replace新增自定义标签](./src/relabel_replace_from_address_to_node_1.png)

我们再看一下指标，**可任选指标**，本次选取`go_memstats_heap_alloc_bytes`展示 , 如图:  

![使用replace新增自定义标签](./src/relabel_replace_from_address_to_node_2.png)

#### 案例5: 慎用 replace改写内部标签

在[案例2: 作用范围是target，而不是job](#案例2: 作用范围是target，而不是job)的基础上，将内置标签`__address__` 的值改写成`biz`标签的值。配置如下 

TODO







#### Relabeling - keep与drop

TODO

#### Relabeling - labelkeep和labeldrop

TODO

#### Relabeling - labelmap

TODO
