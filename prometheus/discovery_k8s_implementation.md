# kubernetes服务发现

## `Discoverer`实际执行者

`prometheus` 定义了 `Discoverer` 接口(*定义文件：`discovery/discovery.go`*)。`prometheus`中。 `Discoverer` 接口只有一个方法` Run(ctx context.Context, up chan<- []*targetgroup.Group) `。 `targets `变化都可以通过监听 `up chan` 获取到。我们将以`kubernetes`为例进行说明。  

**`Discoverer` 接口定义**:  
```go
// Discoverer provides information about target groups. It maintains a set
// of sources from which TargetGroups can originate. Whenever a discovery provider
// detects a potential change, it sends the TargetGroup through its channel.
//
// Discoverer does not know if an actual change happened.
// It does guarantee that it sends the new TargetGroup whenever a change happens.
//
// Discoverers should initially send a full set of all discoverable TargetGroups.

type Discoverer interface {
	// Run hands a channel to the discovery provider (Consul, DNS, etc.) through which
	// it can send updated target groups. It must return when the context is canceled.
	// It should not close the update channel on returning.
	Run(ctx context.Context, up chan<- []*targetgroup.Group)
}

```


