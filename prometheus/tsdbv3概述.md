# TSDB v3 概述


由于`TSDB V2`版本面临上述的诸多问题，`TSDB V3`应运而生。目前`prometheus v2.x`使用就是`TSDB V3`。 2017年,`Prometheus v2`发布之初,将`Prometheus v2`和 `Prometheus 1.8`存储方面进行了比较(详见:[storage:Prometheus v2 vs Prometheus 1.8](https://prometheus.io/blog/2017/11/08/announcing-prometheus-2-0/#storage) ),可见`TSDB V3`对`Prometheus`的性能显著提升了。 本章的重点就是阐述`TSDB V3`。以后没有特别说明`TSDB`指的就是`TSDB V3`

## `TSDB V3`数据库数据文件

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

由上可以看到，文件目录分为三类：`block目录`、`chunks_head`、`wal`

### block目录

目录名为`01BKGV7JBM69T2G1BGBGM6KB12`、`01BKGTZQ1SYQJTR4PB43C8PD98`、  `01BKGTZQ1HHWHV8FBJXW1Y3W0K`就是一个个的`block`。  

默认情况下，`prometheus`以`2个小时`为一个时间窗口，即将`2小时`内产生的数据存  储在一个`block`中。那么监控数据就会被以时间段的形式被拆分成不同的`block`，所  以`prometheus`会产生很多`block`。每个`block`由`ulid`进行标识，例如  `01BKGV7JBM69T2G1BGBGM6KB12`、`01BKGTZQ1SYQJTR4PB43C8PD98`...

每个`block` 目录下包含以下部分：

  - meta.json    元信息  (必须),记录此block的基本信息，例如标识、起始时间时间 戳、终止时间时间戳等
  - tombstones   对数据进行软删除，`prometheus`采用了**标记删除**的策略，将 删除记录保存在`tombstones`中，查询时会根据`tombstones`文件中的删除记录来过 滤已删除的部分.
  - index        索引。
  - chunks       用于保存时序数据。每个`chunks`目录下都有一个或者几个 `chunk`,并且每个`chunk` 最大为`512mb`。超过的部分就会被截断新的`chunk`进行 保存，每个`chunk`以数字编号命名,例如`000001`、`000002`...

### chunks_head

`prometheus`把新采集到的数据存储在内存的`head chunk`。但`chunk`只能写入`120`个样本。如果写满`120`个样本后，才开始进行 `mmap`映射写入写磁盘。然后生成一个空的`chunk` 继续存储新的样本。

通过`mmap`映射写入写磁盘的数据就存储`chunks_head`目录下，`chunks_head`下的数据也以数字编号命名。

### wal 

`prometheus`为了防止应用崩溃而导致内存数据丢失，引入了`WAL`机制。`WAL`采用日志追加模式，数据会先被追加到`WAL`文件中，然后再被刷新到存储引擎中。`WAL`文件被分割成默认大小为 `128MB`的数据段，每个数据段以数字命名，例如 `00000000`、 `00000001`...

同时，我们可以看到`wal`目录下有`checkpoint`目录，那么`checkpoint`是做什么的呢？`WAL`机制是为了备份内存的数据防丢失的。内存中的数据落盘之后，`WAL`备份的数据就没用了。和其他数据库一样，`TSDB`重启之后，也会恢复`WAL`的数据，而且恢复开销大，因此`WAL`备份的数据约简洁约好。所以`prometheus`需要定期清理旧的`wal`数据。对于清理之后剩下的有用的数据，`checkpoint`会将其重新封装成新的数据段，存放在`checkpoint`目录下。

<!-- 
**思考题一** `block`选择`ulid` 作为标识有什么优势吗？为何不选择`uuid`？

`ULID`基于时间戳生成，因此可以按照时间戳进行排序
TODO

**思考题二** `prometheus`既然`wal`是写磁盘，而记录到`block`也是写磁盘。为什么不直接写`block`，而引入`wal`呢？

TODO
**思考题三** 零拷贝技术主要有`mmap`、`sendfile`,为何选择`mmap`,而不使用`sendfile`？

TODO -->


## `TSDB V3`存储流程

存储流程示意图  

![存储流程示意图](./src/tsdb_storage_core.svg)

在上图中，`Head` 块是`TSDB`的内存块，`Block1`、`Block2` ...`BlockN`是持久化磁盘上的文件。同时可以看到，`wal`文件、`chunks_head`也存储于磁盘上。并且说明一个概念:**`chunkRange`**表示`chunk`的记录样本的时间跨度，默认2小时。

`prometheus`中，新数据样本首先会存储于内存的`Head`中。其中`Head`中接收样本的`chunk`称之为`active chunk`。在新数据样本写入内存的`Head`时，会做一次预写日志，将新数据样本写入到`WAL`文件中。  

`chunk`只能写入`120`个样本。**当`active chunk`满了**或者**当前`active chunk`已经持续了`chunkRange`了**，内存的`Head`会创建新的`chunk`来接收新数据样本(*即：新的`active chunk`*)。之前的`active chunk`数据会落盘到`chunks_head`目录中，并通过`mmap`将此部分数据映射到内存。这样`prometheus`就可以根据需要，动态将此部分数据加载到内存了。


`chunks_head`存储的时序时间跨度超过了`chunkRange / 2 * 3`(*注：默认3小时*)，就会将前`chunkRange`时间范围的时序数据压缩到`Block`中。

<!-- 
解读记录：

代码： https://github.com/prometheus/prometheus/blob/main/tsdb/head_append.go
[tsdb/db.go](https://github.com/prometheus/prometheus/blob/main/tsdb/db.go)、
[head](https://github.com/prometheus/prometheus/blob/main/tsdb/head.go)

```go
func (h *Head) compactable() bool {
    if !h.initialized() {
        return false
    }

    return h.MaxTime()-h.MinTime() > h.chunkRange.Load()/2*3
}
``` -->


## WAL

`prometheus`将周期性采集指标，并把指标添加到`active head`，同时`TSDB`通过`WAL`将数据保存到磁盘上进行备份。当出现宕机重启时，就会程读取`WAL`记录，恢复数据。`prometheus`的`WAL`采用**追加写**的方式记录数据，这种顺序写的方式比随机写高效的多。
`prometheus`数据更新都优先会记录内存，也就是会有`WAL`过程，也就意味着数据库的数据都会经历被`wal`文件存储的“经历”。一旦这些更新的数据落盘完成，对应在`WAL`中的备份数据就没用了。如果不清理历史数据、无用数据`WAL`会被“撑爆”的。`WAL`记录是没有经过压缩的，占用空间较大，并且恢复成本页是比较高的。这就是**WAL清理**重要原因。`WAL`清理时，会创建一个`checkpoint`记录清理后的数据状态。

### 编码/数据的组织方法

`prometheus`的`WAL`文件有三种编码类型：`Series records`、`Sample records`、`Tombstone records`。在源码中三种编码的类型[枚举值](https://github.com/prometheus/prometheus/blob/v2.53.0/tsdb/record/record.go#L40)

| 类型   | 编码方式    |
| :-----| :---- | 
| Series records    | 1 (1b)| 
| Sample records    | 2 (1b)| 
| Tombstone records | 3 (1b)|


#### Series Records

由之前的讲解，我们知道`prometheus`的写入模型分三部分：(`labels`、`timestamp`、`value`)。
`labels`是指标唯一标识。但是`labels`一般都很长，每次采集都把`labels`原封不动地存储，那么会造成磁盘空间的极大浪费，并且读写时也会造成很大的`IO`开销。为了解决这问题，`WAL`对`labels`进行了预处理。`labels`会被封装成`Series Record`，写入`wal`文件中,**只写一次**。同时在内存中维护**labels与seriesId**的映射。`seriesId`是`Series Record`类型数据的ID,类型是自增的整形数据。写入数据由(`labels`、`timestamp`、`value`)转换成了(`seriesId`、`timestamp`、`value`)。

**获取`seriesId`的流程示意图**  

![获取`seriesId`的流程示意图](./src/seriesId.drawio.png)

**`Series Records`数据编码**


```
┌───────────────────────────────────────────────────────┐
│ type = 1 <1b> 代码中SeriesRecord类型枚举值为1            │
├───────────────────────────────────────────────────────┤
│ ┌───────────────┬───────────────────────────────────┐ │
│ │ seriesId <8b> │ n= len(labels) labels的对数        │ │
│ ├───────────────┴────────────┬──────────────────────┤ │
│ │ len(label_name1) <uvarint> │ label_name1 <bytes>  │ │
│ ├────────────────────────────├──────────────────────┤ │
│ │ len(label_val1)  <uvarint> | label_val1 <bytes>   │ │
│ ├────────────────────────────┴──────────────────────┤ │
│ │                . . .                              │ │
│ ├────────────────────────────┬──────────────────────┤ │
│ │ len(label_name_n) <uvarint>│ label_name_n <bytes> │ │
│ ├────────────────────────────├──────────────────────┤ │
│ │ len(label_val_n)  <uvarint>| label_val_n <bytes>  │ │
│ └───────────────────────────────────────────────────┘ │
│                  . . .                                │
└───────────────────────────────────────────────────────┘

```

说明：

- `type` :代码中`SeriesRecord`类型枚举值为`1`
- `seriesId` : `SeriesRecord`数据的标识
- `n`:  `SeriesRecord`存储`label`的对数
- 标签`key`的长度、标签`key`的内容以及对应的标签`value`的长度与标签`value`的内容





#### Sample Records




### WAL原理

`WAL`文件被分割成默认大小为 `128MB`的数据段(`segment`)，每个数据段以数字命名，例如 `00000000`、 `00000001`... WAL的写入单位是页(`page`)，每页的大小为`32KB`,数据一次一页地写入磁盘。
每个`WAL`记录都是一个`byte`切片(*注：切片是go语言的特性，本质就是数组，非go语言开发者，可以理解为`byte`数组*)。如何存储`WAL`记录呢？

- 如果一个`WAL`记录大小超过了页(`page`)的大小，这页就会被拆解给更小的子记录，多页存储。
- 如果一个`WAL`记录大小超过了数据段(`segment`)的大小(`128MB`),`prometheus`就会创建更大空间的数据段(`segment`)进行存储。


### WAL清理与CheckPoint


#### WAL清理

TODO


#### CheckPoint

TODO


## chunks_head原理

TODO



## Block持久化

TODO



## 查询与索引


TODO


## 压缩

TODO


## 历史数据清理/数据保留

TODO



## 快照

TODO