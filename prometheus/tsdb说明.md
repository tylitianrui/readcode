# TSDB

`prometheus TSDB`经历了两个主要的版本:  

- `prometheus v1.x`使用`TSDB V2`版本,已经**淘汰**了！！！
- `prometheus v2.x`使用`TSDB V3`版本


## `TSDB V2`

### `TSDB V2`特性

1. **每一个时间序列分别存放到不同的,独立的文件中**

   例如: 上例中四个时间序列会各种存放在四个独立的文件中。例如:

    > - `{__name__="http_request_total",instance="127.0.0.1:9100","job="prom_target",code="200",    method="GET",path="/ping/1"}` 
    > - `{__name__="http_request_total",instance="127.0.0.1:9100","job="prom_target",code="404",    method="POST",path="/pingq"}`
    > - `{__name__="http_request_total",instance="127.0.0.1:9100","job="prom_target",code="404",    method="POST",path="/XXXXX"}`
    > - `{__name__="http_request_total",instance="127.0.0.1:9100","job="prom_target",code="404",method="GET",path="/pingq" }`
  
  
    也就是说`TSDB V2`存在大量的文件，读写时需要保持大量文件处于打开状态,**容易触发系统最大打开文件数**  

    <br/>

2. **时序的最新数据都会缓存到内存，然后批量落盘**
  每一个时序对应**内存**中的独立的`chunk(1KiB)`存储最新的数据，没有`WAL`机制,机器宕机，存储在内存的数据就会丢失。
   <br/>


3. **随机读写**
虽然每个文件是顺序批量写，但`tsdb v2`会读写大量的文件，同时读取这些文件就会产生随机读写的问题。
 <br/>

4. **序列扰动Series Churn**
  `Series Churn` 指的是**一个时间序列集合变得不活跃，即不再接收数据点；取而代之的是出现一组新的活跃序列**  
   如果上例的进程部署在云原生环境中，`instance`用来表示指标来自于哪个实例。如果我们为此服务执行了滚动更新`rolling update`，`instance`就会变化，而产生新的序列。`prometheus`接收不到原序列的指标。除此之外，`Kubernetes`的 `scaling` 也会导致`Series Churn`。  
  
   示意图：

  ```
  series
  ^
  │   . . . . . .
  │   . . . . . .
  │   . . . . . .
  │               . . . . . . .
  │               . . . . . . .
  │               . . . . . . .
  │                             . . . . . .
  │                             . . . . . .
  │                                         . . . . .
  │                                         . . . . .
  │                                         . . . . .
  |
  └── ~ ────────────────────────────────────────────────────>
                                                          time 
  ```
  5.**SSD 写放大的问题**
  时序的最新数据都会缓存到内存，为了节约内存资源，将`chunk`限制为`1KiB`大小。看似节约资源的设计，确带来了新的问题：一旦磁盘是`SSD`硬盘,就可能导致**写放大**问题。 

  **`SSD`硬盘写放大**
  
  TODO

## `TSDB V3`

由于`TSDB V2`版本面临上述的诸多问题，`TSDB V3`应运而生。目前`prometheus v2.x`使用就是`TSDB V3`。 2017年,`Prometheus v2`发布之初,将`Prometheus v2`和 `Prometheus 1.8`存储方面进行了比较(详见:[storage:Prometheus v2 vs Prometheus 1.8](https://prometheus.io/blog/2017/11/08/announcing-prometheus-2-0/#storage) ),可见`TSDB V3`对`Prometheus`的性能显著提升了。 本章的重点就是阐述`TSDB V3`。以后没有特别说明`TSDB`指的就是`TSDB V3`

### `TSDB V3`数据库数据文件

`Prometheus`启动参数`storage.tsdb.path`来指定`tsdb`数据存储位置，默认是`data`目录。我们先看一下`tsdb`数据文件的结构。

目录结构  

```text
./data
├── 01BKGV7JBM69T2G1BGBGM6KB12
│   └── meta.json
├── 01BKGTZQ1SYQJTR4PB43C8PD98
│   ├── chunks
│   │   └── 000001
│   ├── tombstones
│   ├── index
│   └── meta.json
├── 01BKGTZQ1HHWHV8FBJXW1Y3W0K
│   └── meta.json
├── 01BKGV7JC0RY8A6MACW02A2PJD
│   ├── chunks
│   │   └── 000001
│   ├── tombstones
│   ├── index
│   └── meta.json
├── chunks_head
│   └── 000001
└── wal
    ├── 000000002
    └── checkpoint.00000001
        └── 00000000
```

由上可以看到，文件目录分为三类：

1. **block目录**  

目录名为`01BKGV7JBM69T2G1BGBGM6KB12`、`01BKGTZQ1SYQJTR4PB43C8PD98`、`01BKGTZQ1HHWHV8FBJXW1Y3W0K`就是一个个的`block`。  

默认情况下，`prometheus`以`2个小时`为一个时间窗口，即将`2小时`内产生的数据存储在一个`block`中。那么监控数据就会被以时间段的形式被拆分成不同的`block`，所以`prometheus`会产生很多`block`。每个`block`由`ulid`进行标识，例如`01BKGV7JBM69T2G1BGBGM6KB12`、`01BKGTZQ1SYQJTR4PB43C8PD98`...

每个`block` 目录下包含以下部分：

- meta.json    元信息  (必须),记录此block的基本信息，例如标识、起始时间时间戳、终止时间时间戳等
- tombstones   对数据进行软删除，`prometheus`采用了**标记删除**的策略，将删除记录保存在`tombstones`中，查询时会根据`tombstones`文件中的删除记录来过滤已删除的部分.
- index        索引。
- chunks       用于保存时序数据。每个`chunks`目录下都有一个或者几个`chunk`,并且每个`chunk` 最大为`512mb`。超过的部分就会被截断新的`chunk`进行保存，每个`chunk`以数字编号命名,例如`000001`、`000002`...

2. **chunks_head**
`prometheus`把新采集到的数据存储在内存的`head chunk`。但`chunk`只能写入`120`个样本。如果写满`120`个样本后，才开始进行 `mmap`映射写入写磁盘。然后生成一个空的`chunk` 继续存储新的样本。
通过`mmap`映射写入写磁盘的数据就存储`chunks_head`目录下，`chunks_head`下的数据也以数字编号命名。

3. **wal**

`prometheus` 为了防止应用崩溃而导致内存数据丢失，引入了`WAL`机制。`WAL`采用日志追加模式，数据会先被追加到`WAL`文件中，然后再被刷新到存储引擎中。`WAL`文件被分割成默认大小为 `128MB`的数据段，每个数据段以数字命名，例如 `00000000`、 `00000001`...

<!-- 
**思考题一** `block`选择`ulid` 作为标识有什么优势吗？为何不选择`uuid`？

`ULID`基于时间戳生成，因此可以按照时间戳进行排序
TODO

**思考题二** `prometheus`既然`wal`是写磁盘，而记录到`block`也是写磁盘。为什么不直接写`block`，而引入`wal`呢？

TODO
**思考题三** 零拷贝技术主要有`mmap`、`sendfile`,为何选择`mmap`,而不使用`sendfile`？

TODO -->

