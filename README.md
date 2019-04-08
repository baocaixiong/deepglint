### redis通讯协议简单实现

- 实现了对string，int，error，array，bulk string的解析
- 当前支持使用管道方法调用
    ```go
    echo -e ':1000\r\n' | go run protocol.go
    echo -e '*2\r\n$3\r\nbar\r\n$5\r\nhello\r\n' | go run protocol.go
    ```