# BGPServer

**实时获取全球BGP路由软件**

### 运行命令参数说明： 
#### -c 指定config.ini 文件路径（相对路径即可）<br>
#### -l 指定需要获得的国家路由，例如：CN、US、HK，另外有两个特定参数：NCN（所有非CN路由）； ALL（全球BGP路由） <br>

#### 例如1：./bgp -c ./config.ini -l ALL （全球BGP）<br>
#### 例如2：./bgp -c ./config.ini -l US （美国BGP路由）<br>
##### 例如3：./bgp -c ./config.ini -l CN （中国BGP路由）<br>
#### 例如4：./bgp -c ./config.ini -l NCN （中国以外的BGP路由）<br>

#### config.ini配置文件只需设置 服务器：routerid、peer、nexthop、peer对端IP、peer对端ASN、updateSource <br>

#### 建议在能科学上网的环境下运行 否则不一定保证能接收到实时路由 <br>

#### BGP 通信端口为 TCP 179 需要开放此端口通信 <br>

#### 使用方法：
##### 1、下载releases中的：bgp，ChinaBGPZip.gob，config.ini，map.json.gz 四个文件到同一个目录下，比如/root/bgp
##### 2、赋予bgp文件执行权限：sudo chmod +x bgp
##### 3、前端运行直接复制上面的列如模式即可
##### 4、后端运行请参考下面的方式：
###### 4.1、sudo apt-get update
###### 4.2、sudo apt-get install screen
###### 4.3、sudo screen -S bgp_session -d -m ./bgp -c ./config.ini -l CN
###### 4.4、查看运行状态：sudo screen -r bgp_session
###### 4.5、先按下 Ctrl + A，然后按下 Ctrl + D，这将从 screen 会话中断开连接，但会保持会话处于后台运行状态。
###### 4.6、终止会话直接按下：Ctrl + C

##### *仅用于学习目的*
