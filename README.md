-----------+           +----------------+           +----------------+
|   transport    |  <----->  |      hub       |  <----->  |    storage     |
| (tcp/ws/other) |    I/O    | (chat core)    |  events   | (postgres/redis)|
+----------------+           +----------------+           +----------------+
^                            ^   ^                        ^
|                            |   |
client sockets                register messages              pub/sub
(telnet / browser)                 / broadcast                  (分布式)