# golang_chatserver
# 聊天室服务器

.<br />├── go.mod<br />├── go.sum<br />├── hiface -------网络层接口<br />│    ├── iconnection.go -------连接的接口<br />│    ├── imessage.go ----------消息的接口<br />│    └── iserver.go -------------服务器的接口<br />├── hnet  --------网络层抽象<br />│    ├── netconnection.go -------单个连接的抽象<br />│    ├── netproto.go -------------消息和拆封包对象的抽象<br />│    ├── netserver.go -------------服务器的抽象<br />│    └── networkthread.go -------任务线程和异步工作池的抽象<br />├── hzhgagaga -------可执行文件<br />├── main.go -------服务器主函数<br />├── README.md<br />└── server -------业务层（MVC中的MC模式）<br />    ├── core -------公用模块<br />    ├── model --------M数据层，需要操作的数据或信息<br />    ├── msgwork --------C控制层，处理消息逻辑<br />    ├── pb ---------protobuf协议文件存放地<br />    ├── siface --------接口<br />    └── theworld.go ----------业务世界管理对象<br />
<br />

