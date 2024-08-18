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
