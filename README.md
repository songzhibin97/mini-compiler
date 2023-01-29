# Mini-compiler

以 https://github.com/songzhibin97/mini-interpreter 为基础,通过栈来实现简单的编译器

```
.
├── README.md
├── code // 指令定义
│ ├── code.go
│ └── code_test.go
├── compiler // 编译器
│ ├── compiler.go
│ ├── compiler_test.go
│ ├── func.go
│ ├── symbol_table.go
│ └── symbol_table_test.go
├── go.mod
├── go.sum
├── main.go
├── repl
│ └── repl.go
└── vm // 虚拟机
    ├── frame.go
    ├── vm.go
    └── vm_test.go

```

## Demo



```shell
# go run main.go

Welcome to Mini-compiler
>>>print("hello world")
hello world
>>>
```


```shell
>>>var i = 1
>>>var b = true
>>>var s = "string"
>>>var array = [1,2,3]
>>>var mp = {1:1,2:2}
>>>print(i,b,s,array,mp)
1
true
string
[1, 2, 3]
{1:1, 2:2}

>>>print(array[1])
2

>>>print(mp[2])
2

>>>func add(a,b){return a + b}
>>>print(add(1,2))
3

>>>print( 1 + 2 * 3 )
7

>>>func max(a, b) {if (a > b) { return a } return b }
>>>print(max(1,2))
2
>>>print(max(2,1))
2

>>>func max(a, b) {if (a > b) { return a } else { return b }}
>>>print(max(1,2))
2
>>>print(max(2,1))
2

>>>print(!true)
false

>>>print(-10)
-10
```

## benchmark
```
fibonacci 35
BenchmarkInterpreter
BenchmarkInterpreter-12    	       1	61665154927 ns/op
BenchmarkCompiler
BenchmarkCompiler-12       	       1	12110596378 ns/op

fibonacci 20
BenchmarkInterpreter
BenchmarkInterpreter-12    	      25	  43752838 ns/op
BenchmarkCompiler
BenchmarkCompiler-12       	     134	   9115634 ns/op

fibonacci 15
BenchmarkInterpreter
BenchmarkInterpreter-12    	     272	   3955469 ns/op
BenchmarkCompiler
BenchmarkCompiler-12       	    1203	    993288 ns/op
```
