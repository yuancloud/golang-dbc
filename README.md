# Golang-DBC

一个使用golang实现的DBC信号接收器和DBC解析器, 开箱即用

## 使用

运行方式

```bash
go build -o golang-dbc main.go

# 指定需要连接的IP和Port，以及当前使用的DBC文件
./golang-dbc -ip=192.168.100.9 -port=5500 -dbc=dbc/xxx.dbc
```

通过接口查看DBC信号
```bash
curl localhost:8090/dbc-vars
```

样例数据

```json
{
  "code": 0,
  "data": {
    "sig1": 1.0,
    "sig2": 24.4,
    "sig3": 12.28,
    "sig5": 24.12
  },
  "msg": "success"
}
```
