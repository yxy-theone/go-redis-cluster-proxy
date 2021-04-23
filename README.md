# goRedisClusterProxy

#### 介绍
使用golang实现一个rediscluster操作代理接口
方便lua, erlang等服务调用

#### 
简要说明


#### 监听端口

1. Addr: ":8282"
2. 路由： http://127.0.0.1:8282/goRedis  【POST】

#### redis cluster 连接

initRedis()


#### 操作说明

1. 参数为json数组，元素1为命令，后续为key-value等参数，支持大部分命令。
2. 若返回值不是string的，需要在typeof方法中配置。
3. 批量执行方法的返回结果有待封装。 


#### 操作示例

1. 单命令
{"data": ["set", "dfg","5678"]}
{"data": ["get", "dfg"]}

2. 批量执行
{"data": ["batch", "[\"set\",\"batch_test3\",\"1\"]","[\"set\",\"batch_test4\",\"2\"]"]}

#### 后续

1. go新手，欢迎优化pr, thanks

