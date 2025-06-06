# 6.5840  2024

大名鼎鼎的6.824（现6.5840）

---

## intro

有很多分布式系统，本课程大多是关于’基础服务设施‘

例如：大型网站的存储，MapReduce，p2p共享

构建分布式系统很难

1. 并发问题
2. 复杂的交互
3. 很难实现高性能
4. 系统的部分瘫痪

为什么还要构建分布式？

1. 通过并行计算，提升服务能力
2. 通过复制容忍故障
3. 匹配物理设备的分布，例如传感器（物理设备的多样性，分散性，使得集中式系统难以应付）
4. 通过隔离提高安全性

## MAIN TOPICS

大型应用程序往往构建在可靠的分布式系统之上，所以这是一门关于应用程序基础设施的课程

1. 存储
2. 通信
3. 计算

分布式系统的目标！隐藏极其复杂的分布式系统的细节，对上层应用提供友好的接口！

### topic：fault tolerance

成百上千的服务器，巨大的网络结构，总会产生故障，我们希望分布式系统能对上层应用隐藏这些故障，即便部分故障，也能继续服务。

### topic：consistency

一致性。数据符合预期，例如read(x) 的结果是最近一次 write(x) 的值。一致性对于多副本的系统很难实现

### topic：performance

高性能，即可扩展的吞吐量，可扩展的cpu，ram，disk，net。当节点数越多，扩展则越难，主要难点如下

1. 负载不均衡
2. 短板效应
3. N增大，则很多过程会变慢，例如初始化，交互

### topic：tradeoffs

权衡！

容错性，一致性，高性能。这三者不可兼得

容错和一致性 需要结点的通信，而通信总是会慢，且随着系统规模增加而变慢

许多分布式的设计牺牲了一致性，来换取速度

例如：读取的数据不一定是最新写入的数据

这对于程序员和用户来说可能都是痛苦

在本课程中将看见很多不同的关于 **一致性/高性能** 的设计

### topic：implementation

rpc，线程，并发控制，配置

## 举个例子

MapReduce（MR）！MR是个很好的例子来介绍本课程的主题

### MR 概览

**场景**：在TB级别 的数据上进行 几个小时的计算

**例如**：建立索引，排序，分析网络结构

只有上千台机器才可能完成的任务！

而且上层应用程序员不是分布式系统专家！

**目标**：是的非分布式专业的程序员也能很好的使用，程序员只用定义Map和Reduce的内容即可！而这些一般是简短的代码

### 以单词计数来举例 MR的大致工作流程

```
Input1 -> Map -> a,1 b,1
Input2 -> Map ->     b,1
Input3 -> Map -> a,1     c,1
                  |   |   |
                  |   |   -> Reduce -> c,1
                  |   -----> Reduce -> b,2
                  ---------> Reduce -> a,2
```

1. input被分为M片
2. MR调用map方法，分别处理M片数据，并输出<k,v>键值对（中间计算结果），每次调用map都是一个任务
3. map阶段结束，MR收集所有中间计算结果，并调用Reduce，来对单词分别进行统计
4. 最终收集所有reduce的输出作为结果

### MR有很好扩展性

即，N 个结点能带来 N倍的吞吐量！

因为，map操作不需要节点间交互，所以map操作可以在每个节点上并行执行，reduce同理。

因此，更多的结点 -> 更大的吞吐量

ps：MR优秀的扩展性不是绝对的，MR也可能会随着扩展而在某些过程中产生性能下降的情况，只是相对于其他分布式系统而言MR扩展性更好

### MR隐藏了非常多细节

发送 M R 请求，跟踪任务完成情况，中间数据处理，负载均衡，宕机恢复……

### 为了获得这些优点，MapReduce强调如下规则

节点没有状态 & 节点间极少的交互
只有一个数据流模式， 即Map/Reduce
没有实时或流式处理。

### 一些MR的细节

详情见MR的论文中的Figure 1
