# FOR

```go
for i := 1; i < 10; i++ {
    // code
}


for ; i < 10; {
    // code
}


// go's while, go has no while syntax
for i < 10 {
    // code
}


for ;; {
    // forever loop
}

for {
    // forever loop
}
```

# IF

```go
if i < 10 {
    // code
}


// Like for, you can execute a short statement before the condition
// Variables declared by the statement are only 
// in scope until the end of the if.
if i := 1; i < 10 {
    // code
}
```

# Switch

```go
// 和for，if一样，switch也可以在开头执行一些短语
switch os := runtime.GOOS; os {
    case "darwin":
        fmt.Println("OS X.")
    case "linux":
        fmt.Println("Linux.")
    default:
        // freebsd, openbsd,
        // plan9, windows...
        fmt.Printf("%s.\n", os)
    }


// 与c++，java不同，switch命中多个case只会执行第一个命中的
i := 2
switch {
    case i > 1:
        fmt.Println("i > 1")
    case i >= 2:
        fmt.Println("i >= 1")
}


// fallthrough可以让下一个case执行，无需检查条件
switch {
    case i > 1:
        fmt.Println("i > 1")
        fallthrough
    case i >= 2:
        fmt.Println("i >= 1")
}
```

# Defer

```go
func main() {
    // defer 的语句直到return时才执行 
    defer fmt.Println("world")

    fmt.Println("hello")
}
```

判断下面的main函数输出什么

```go
package main

import "fmt"

func namedReturn() (result int) {
    result = 1
    defer func() {
        result++
        fmt.Println("defer 语句执行，result 变为", result)
    }()
    return result
}

func main() {
    res := namedReturn()
    fmt.Println("函数返回值为", res)  // 2
}
```

判断下面的main函数输出什么

```go
package main

import "fmt"

func unnamedReturn() int {
    num := 1
    defer func() {
        num++
        fmt.Println("defer 语句执行，num 变为", num)
    }()
    return num
}

func main() {
    res := unnamedReturn()
    fmt.Println("函数返回值为", res) // 1
}
```

上述情况中，函数返回值是否有名，会有不同影响

多个defer，回记录到defer栈中，按照LIFO顺序执行

```go
func main() {
    fmt.Println("counting")

    for i := 0; i < 10; i++ {
        defer fmt.Println(i)
    }

    fmt.Println("done")
}
```

# Pointer

指针，类似C，

- *T就是指向T类型的指针

- 指针的值为0则代表nil（空）

- &为取地址符号，*为取值符号

- 和C不同的是，go没有指针计算语法

# Struct

go没有class，只有struct，但是也能实现面向对象。

```go
type Vertex struct {
    X int
    Y int
}

func main() {
    v := Vertex{1, 2}
    v.X = 4
    fmt.Println(v) // output is "{4, 2}"
    fmt.Println(Vertex{1, 2})  // output is "{1, 2}"

    // 指针
    p := &v
    p.X = 1e9    // 标准写法是(*p).X = 1e9, 但是可以简写
    fmt.Println(v)
}
```
