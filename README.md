# goX
Write some ADVANCED code implementations in Go ...... to deepen understanding and usage of Golang

## Update-Timeline

### 1. LRU With Young-Old Region

代码于 `goX/lrux` 中

在 Innodb 中采用了区分 Young 和 Old（Non-Young） 区的[LRU 优化方案](https://dev.mysql.com/doc/refman/8.4/en/innodb-buffer-pool.html)，本实践简单实现了双区域的划分。



