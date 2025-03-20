# Key/Value Server

环境：单机

操作语义：exactly-once

一致性要求：可线性化（Linearizable）





### Key/value server with no network failures ([easy](http://nil.csail.mit.edu/6.5840/2024/labs/guidance.html))

很简单，锁住对于map的操作就行了。
