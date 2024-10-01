# goX
Write some ADVANCED code implementations in Go ...... to deepen understanding and usage of Golang

## Update-Timeline

### 1. LRU With Young-Old Region

代码于 `goX/lrux` 中

在 Innodb 中采用了区分 Young 和 Old（Non-Young） 区的[LRU 优化方案](https://dev.mysql.com/doc/refman/8.4/en/innodb-buffer-pool.html)，本实践简单实现了双区域的划分。


### 2. Web Framework based on net/http

代码于 `goX/bttp` 中

参考了 `gin` 的功能设计，以及 [Gee](https://geektutu.com/post/gee.html) 这篇文章提供的代码和思路演进。

**TODO-LIST**

- [x] 支持路由注册、HandlerFunc 绑定
- [x] 基础前缀树路由
- [x] 支持动态路由
- [ ] 支持路由组
- [ ] 支持 Middleware 设计