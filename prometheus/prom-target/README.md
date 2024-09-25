# prom-target

## 启动命令

```shell
make  run
```

## promethus 采集metrics

```shell
curl  http://127.0.0.1:9100/metrics
```

## 对外服务

### 状态码 200 

```shell
curl  http://127.0.0.1:8520/ping/123
```

### 状态码 400

```shell
curl  http://127.0.0.1:8520/ping/tyltr
```
