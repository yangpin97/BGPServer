# BGPserver
a live BGP server peer FOR any country  实时获取全球BGP 路由软件
###    编辑config.ini 配置文件 设定相关参数
###    运行命令 参数 -c 为指定config.ini 文件路径（相对路径即可）
###    -l 参数 指定需要获得的国家路由 例如CN US 特定两个参数NCN 所有非CN路由 ALL 为全球BGP路由 
. 例如 bgp -c ./config.ini -l ALL  为全球BGP bgp -c ./config.ini -l US 为美国路由 无需区分大小写
. config.ini 配置文件只需简单设置 服务器routerid peer nexthop 和peer端IP 和peer 端ASN即可
##### 程序运行需要两个文件 map.json.gz chinaBGPZip.gob 务必放置于bgp文件同一目录下 
