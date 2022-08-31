# ztil
工具，方便平时用于调试

```
go install ./

ztil -h
```

## 功能
```
COMMANDS:
   crc32, crc         crc32 <key>
   kafka_receive, kr  receive kafka message
   kafka_send, ks     Send kafka message
   mysql, sql         select sql
   redis, r           redis
   v50
   help, h            Shows a list of commands or help for one command
```
- ks kafka 生产者
- kr kafka 消费者

### kafka_send, ks
Send kafka message
```
USAGE:
ks <topic> <value> [url]
```
