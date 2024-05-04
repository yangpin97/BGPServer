# BGPServer

**实时获取全球BGP路由系统，路由采集信息以中国地区为优化any cast类型的路由**

###    编辑config.ini 配置文件 设定相关参数<br>

###    运行命令 参数 -c 为指定config.ini 文件路径（相对路径即可）<br>

###    -l 参数 指定需要获得的国家路由，例如：CN、US、HK，特定两个参数：NCN（所有非CN路由）； ALL（全球BGP路由） <br>

#### 例如1：./bgp -c ./config.ini -l ALL （全球BGP）<br>
#### 例如2：./bgp -c ./config.ini -l US （美国BGP路由）<br>
#### 例如3：./bgp -c ./config.ini -l CN （中国BGP路由）<br>
#### 例如4：./bgp -c ./config.ini -l NCN （中国以外的BGP路由）<br>

##### config.ini配置文件只需简单设置 服务器routerid、peer、nexthop、peer对端IP、peer对端ASN、updateSource 即可 <br>

##### 程序运行只需要两个文件 map.json.gz、chinaBGPZip.gob，请务必放置于bgp文件同一目录下 <br>

##### 建议在能科学上网的环境下运行 否则不一定保证能接收到实时路由 <br>

##### BGP 通信端口为 TCP 179 需要开放此端口通信 <br>

##### *仅用于学习目的*
