# goX
Write some ADVANCED code implementations in Go ...... to deepen understanding and usage of Golang

## Update-Timeline

### 1. LRU With Young-Old Region

代码于 `goX/lrux` 中

在 Innodb 中采用了区分 Young 和 Old（Non-Young） 区的 [LRU 优化方案](https://dev.mysql.com/doc/refman/8.4/en/innodb-buffer-pool.html)，本实践简单实现了双区域的划分。


### 2. Web Framework Based On net/http

代码于 `goX/bttp` 中

参考了 `gin` 的功能设计，以及 [Gee](https://geektutu.com/post/gee.html) 这篇文章提供的代码和思路演进。

**TODO-LIST**

- [x] 支持路由注册、HandlerFunc 绑定
- [x] 基础前缀树路由
- [x] 支持动态路由
- [x] 支持路由组
- [x] 支持 Middleware 设计


### 3. HashMap Based On Consistent Hash

代码于 `goX/cachex/consistent_hash` 中

实现了基础的一致性哈希功能。

**TODO-LIST**

- [x] 支持指定节点的虚拟节点数量
- [x] 支持用户自定义哈希函数
- [ ] 支持 kv 过期
- [ ] 支持 下线节点/上线节点 时触发数据自动迁移

### 4. RPC Framework Designed Like `net/rpc`

代码于 `goX/brpc` 中

代码和架构风格参考 `net/rpc` 库源码与 [GeeRPC](https://geektutu.com/post/geerpc.html)，重写实现了若干增强方法。
当前实现存在 bug，未对代码进行完善，等日后填坑。。。

**TODO-LIST**

- [x] 多应用层编解码协议协商
- [x] 客户端/服务端支持并发、异步调用
- [x] 加入客户端/服务端的超时控制机制
- [ ] 客户端自动 failover
- [ ] 对接注册发现中心

### 5. Hash Map 

代码于 `goX/bmap` 中

**TODO-LIST**

- [ ] 并发检测
- [ ] 自动扩容
- [ ] hash-slot 仿 golang 原生的 bucket 实现
- [ ] 细粒度锁
- [ ] 泛型支持