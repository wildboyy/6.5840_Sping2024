# Lecture 2: Threads and RPC

### 主题：

分布式系统的实现，实验中的go编程，go线程，网络爬虫，go rpc

### 为什么选择 Go 语言？

- 对协程（轻量级线程）有很好的支持
- 具备便捷的远程过程调用（RPC）机制
- 类型安全（即减少出现因类型不匹配导致的问题）且内存安全
- 拥有垃圾回收机制（避免释放内存后仍使用的问题）
  - 协程与垃圾回收机制的结合尤其有吸引力！
- 不太复杂
- Go 语言常用于分布式系统中

### 有没有线程的替代方案？

有，

编写在单线程中明确交织执行多个活动的代码。 这通常被称为“事件驱动”编程。 

为每个活动（例如每个客户端请求）维护一个状态表。 有一个“事件”循环，它会： - 检查每个活动是否有新的输入（例如，来自服务器的响应到达）； - 为每个活动执行下一步操作； - 更新状态。 

事件驱动编程可以实现 I/O 并发， 并且消除了线程开销（这种开销可能相当大）， 但无法利用多核处理器实现加速， 而且编程过程很痛苦。 

### 多线程挑战

共享数据安全

“竞争”：多个线程同时操作同一块内存

解决竞争的方法：

1. 锁
2. 避免共享可变数据

线程间如何协作？

用 Go channels ，sync.Cond, sync.WaitGroup

死锁？

### Crawler(爬虫)

**爬虫**：递归的去爬网页数据。随便给它一个起始网站，它递归地从网页内的连接去爬其他网站的数据。一个url只被爬一次，且不能陷入url循环中。

**挑战**：

- 充分利用并发

- 一个URL只爬一次

- 结束条件

课程代码 crawler.go 给了三种实现方式

- 单线程

- 并发，通过共享内存协作

- 并发，通过channel协作

## ConcurrentChannel crawler

### Go channel

go通道，非常类似于OS中的PV原语

```go
// channel是个对象
// a channel lets one thread send an object 
// to another thread
ch := make(chan int)

// the sender waits until some goroutine receives
ch <- x

// a receiver waits until some goroutine sends
y := <- ch

// y 循环从通道中获取对象，直到通道关闭，则退出循环
for y := range ch
```

channel的发送和接受只**消耗不到1ms的时间**

## 锁 vs channel

什么情况用锁 + 共享内存？

什么情况用通道？

**取决于程序员的思维模式：**

- ​**状态管理** → 共享数据 + 锁（`sync.Mutex`/`sync.RWMutex`）
- ​**通信协调** → 通道（`channel`）

## Remote procedure call（RPC）

### rpc如何解决错误？

#### 1、最大努力（at-least-once）

call 方法等待回复，一段时间无回复则重新发送，循环这个过程几次，最终放弃并返回错误

#### 2、最多一次 (at-most-once )

服务端（被调用者）会记录请求id，并且避免重复执行。

go rpc就是这种

#### 3、精确一次（exactly once）

无限重复call + 重复执行检测 + 容灾
