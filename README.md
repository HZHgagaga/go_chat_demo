CREATE TABLE IF NOT EXISTS `chat_msg`(<br />   `chat_id` INT UNSIGNED AUTO_INCREMENT,<br />   `chat_time` VARCHAR(100) NOT NULL,<br />   `chat_name` VARCHAR(100) NOT NULL,<br />   `chat_data` TEXT,<br />   PRIMARY KEY ( `chat_id` )<br />)ENGINE=InnoDB DEFAULT CHARSET=utf8;<br />
<br />
<br />数据包使用简单的TLV格式<br />T->协议类型(4字节)，L包体长度(4字节)，V是protobuf打包的数据内容<br />
<br />hnet为网络层的内容 
- 一个连接各一个读写协程循环读循环写
- 读协程读出来的数据放入业务处理channel中
- 业务处理单协程循环执行channel的任务
- 提供了一个协程池，处理异步IO，协程数与CPU核数相等，取随机数负载协程

<br />hiface为网络层提供的接口

<br />server为业务层内容
- 业务层在msgwork文件夹下添加协议处理结构体
- 服务器启动时，遍历已添加的协议处理结构体，如果protobuf文件有声明相应的ID，用反射将各结构体方法与数据包的消息ID映射起来
- 有包发过来的时候，通过映射map执行相应协议处理方法
- 业务层处理完数据，将回复封包打入发送协程的channel中
- 业务层遇到耗时任务应该将其放入异步池处理

<br />server/msgwork为处理结构体的内容<br />
<br />server/the_world管理所有业务层的相关信息<br />
<br />server/player为玩家的抽象<br />
<br />server/iserver为业务层的接口<br />
<br />业务功能：<br />广播聊天 聊天内容存入mysql<br />查看当前在线的用户名<br />一对一聊天 通过用户名发送私密聊天消息
