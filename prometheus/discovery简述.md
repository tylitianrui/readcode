# 服务发现简述

作为监控系统，`prometheus`首先要解决的就是"`要监控谁`"的问题。 
静态配置(`static_config`)将被监控对象地址“写死”在配置文件中，这比较适合`targets`稳定不变的场景。配置方式在[prometheus功能介绍-服务发现](prometheus功能介绍.md#服务发现)章节已经做了说明。  
在云原生体系下，各种资源是动态分配的，是随着需求规模的变化而变化的。这也意味着没有固定的监控目标。这也就无法继续使用静态配置(`static_config`)方式处理被监控对象了。那怎么解决这问题呢？目前最常规的策略就是**服务注册与服务发现**。被监控对象讲自身信息注册到**注册中心**。`prometheus`实现服务发现功能，在**注册中心**获取最新的监控目标的信息。

目前版本的`prometheus`(`v2.53`) 实现了多种种服务发现功能，具体可见[prometheus的配置文档](https://prometheus.io/docs/prometheus/2.53/configuration/configuration/)。最常见的是以下几种：

- **static_config** 静态配置文件
- **file_sd_config** 基于本地文件的服务发现。`target`发生变化时会更新到文件中;`prometheus`会定时从文件中读取最新的`target`信息。这种方式不需要依赖于第三方的注册中心，是最简单的服务发现方式。
- **kubernetes_sd_config** 基于`kubernetes api-server` 获取

## static_config

配置文件：**prometheus.yml**

```yaml
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

  scrape_configs:
    - job_name: "prometheus"
      metrics_path: "/metrics"
      static_configs:
        - targets: ["127.0.0.1:9090"]
```

如图：  
![prometheus_static_config](./src/prometheus_static_config.png)


## file_sd_config


配置文件：**prometheus.yml**

```yaml
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

scrape_configs:
  - job_name: "prometheus_file_sd_configs"
    file_sd_configs:
      - files: ["./file_sd_config/*.yml"]
        refresh_interval: 15s
```



配置文件：**./file_sd_config/prometheus_targets.yml**

```yaml
- targets:
  - '127.0.0.1:9090'

  labels:
    times: "2"
```

上述配置，`prometheus`会定期读取文件中的内容读取文件内容。当文件中定义的内容发生变化时，不需要重启`prometheus`,即可加载`targets`实现监控。

- `files` 读取文件的目录。本处为`./file_sd_config/*.yml`，相对于`prometheus.yml`文件。
- `refresh_interval`设置读取`files`的时间间隔，本处为`15s`


如图：  
![prometheus_file_sd_configs](./src/prometheus_file_sd_configs.png)