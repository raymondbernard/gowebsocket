# gowebsocket
Simple server (and client) handling communication over web sockets. Messages are broadcasted over websockets. 

# How to use it

- First start up a server:

``` go
ip := "localhost"
port := "3999"
server := gowebsocket.New(ip, port)
server.Start() 
```

- Any client can connect to it by

``` go
c := gowebsocket.NewClient(ip,port) 
```

- Messages can be sent/received by using _.Send()_ and _.Receive()_ methods: 

``` go

c.Send("Hello there\n")
recv := c.Receive()
fmt.Println("Received: ", recv)

```

See _test/websocket-test.go_ for more info! 

# Credit
Implementation is heavily based on http://gary.beagledreams.com/page/go-websocket-chat.html. 

