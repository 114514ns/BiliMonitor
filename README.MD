# BiliMonitor
~~视奸你喜欢的up主~~
# Features
- 定时记录关注列表内的up主的粉丝数
- 记录直播间内弹幕，礼物流水
- 当主播开始直播/发布动态时发送邮件/Server酱/QQ提醒
- 记录主播大航海成员
- 在web上查看记录到的数据
- 备份新稿件到Alist
- 联动Mikufans录播姬
  - 自动上传录播文件
  - 上传时将flv格式转化为mp4
  - 单独上传直播音频
- 可选MySQL和SQLite数据库
- 备份(爬取)用户动态和评论
- 下载单个视频/分P视频/播放列表
- 自动下载收藏夹内的视频（使用场景：想要下载一个视频，或者感觉这个视频会被下架掉，只要点收藏就可以备份/快速获得视频链接）
- 爬虫：
  - 爬取分区内主播粉丝量，大航海成员，粉丝团成员
  - 支持分布式
  - 支持设置代理
- 一个简易的直播间页面，可以观看直播和发送弹幕
    
scripts目录下有一些数据整理的python脚本

# Communication

任何问题和建议欢迎在issue或discussion内提出

# Download
https://github.com/114514ns/BiliMonitor/actions
<br/>
当大部分功能完成后会发布release，现在先用actions

# Screenshots

![](https://example.com/undefined20250305001302952.png)
![](https://example.com/undefined20250305001138294.png)
![](https://example.com/undefined20250302233820130.png)
![](https://example.com//img/20250502170730.png)
![](https://example.com/undefined20250625001439153.png)
![](https://example.com//img/20250212235435.png)
![](https://example.com/undefined20250220222834082.png)
![](https://example.com/undefined20250302233902605.png)
![](https://example.com/undefined20250302233924636.png)
![](https://example.com/undefined20250511002801489.png)
![](https://example.com/undefined20250627001835961.png)
![](https://example.com/undefined20250625001246302.png)
![](https://example.com/undefined20250627001748657.png)


# Known issues
- 暂不支持mips64架构
# Configuration

```json5
{
  "RefreshFollowingsDelay": "30m", //刷新关注列表间隔
  "User": "98618805", //你的uid，列表内的所有up主会被定时记录粉丝量，需要设置关注列表允许公开访问
  "CommonDelay": "30m", //记录粉丝量的间隔
  "SpecialList": [], //特别关注列表，列表内的up主发布的动态和稿件会被推送，类型为int
  "SpecialDelay": "2m", //刷新特别关注列表内up主动态的间隔
  "Cookie": "", //b站cookie，f12随便找个请求然后复制cookie粘贴过来
  "RefreshToken": "", //RefreshToken，需要正确填写，否则无法刷新Cookie
  "EnableEmail": false, //启用邮件推送，目前使用resend api实现发送邮件
  "ResendToken": "re_TLeNcEDu_Ht8QFPBRPH6JyKZjfnxmztwB", //申请的token,一天免费100封邮件，你可以暂时用我的来测试。
  "FromMail": "bili@ikun.dev", //发件邮箱
  "ToMail": [], //收件邮箱
  "EnableQQBot": false, //启用OneBot推送
  "ReportTo": [
    "10001" //要推送到的qq账号
  ],
  "BackServer": "http://127.0.0.1:3090", //OneBot后端
  "EnableServerPush": false, //启用Server酱推送，参考https://sc3.ft07.com/接入
  "ServerPushKey": "", //Server酱key
  "Tracing": ["544853"], //要监控的直播间，类型为string[]，不能是短号
  "EnableAlist": false, //启用alist相关功能
  "AlistServer": "http://192.168.10.16:5244/", //alist服务器，注意url最后的斜杠
  "AlistUser": "admin", //alist用户名
  "AlistPass": "password", //alist密码
  "AlistPath": "移动/Bili/", //要保存到的路径
  "EnableLiveBackup": true, //启用直播备份
  "MikuPath": "/mnt/c/Users/Administrator/Desktop/Record/", //录播姬路径
  "EnableSQLite": true, //启用SQLite数据库
  "SQLitePath": "database.db", //sqlite数据库名字
  "EnableMySQL": false, //启用MySQL数据库
  "SQLName": "bili", //数据库名
  "SQLUser": "bili", //用户名
  "SQLPass": "CERtxrAH2zaik7b2", //密码
  "SQLServer": "127.0.0.1:3306", //数据库地址
  "CodeToMP4": true, //将录播姬的flv格式转为mp4格式
  "SplitAudio": true, //上传录播时另外上传一份直播的声音
  "EnableCollectionMonitor": true, //启用收藏夹监控
  "Mode": "Master", //分布式的模式 取值为 Master/Slaver
  "Slaves": [""] ,//子节点列表
  "TraceArea": true, //是否启用分区监控
  "BlackTracing": ["21013884"], //房间黑名单，注意是房间号
  "HTTPProxy": "http://fg27msTTyo:PZ8u9Pr2oz@[172.23.65.46]:8080", //HTTP代理
  "QueryProxy": ""// 查询代理，不需要Cookie的查询接口会走这个代理
}
```
大部分功能都需要Cookie，所以请先填写好Cookie。
<br>
## 粉丝量监控
```json5
{
  "User": "98618805",
}
```
这里改成你的uid，然后允许公开访问你的关注列表即可。

程序会每隔指定时间记录关注列表内up主粉丝数

目前没有前端界面显示粉丝量的趋势，只存储在数据库里，后面会加。

## 收藏夹监控
~~~json5

{
  "EnableCollectionMonitor": true
}
~~~
被监控的用户就是上面设置的那个User。

新建一个名字为Monitor的收藏夹，并设为可公开访问，当收藏夹内出现新的视频时可自动下载并发出推送。
## 动态监控
```json5
{
    "SpecialList": [12434430], //特别关注列表，列表内的up主发布的动态和稿件会被推送，类型为int
    "SpecialDelay": "2m", //刷新特别关注列表内up主动态的间隔
}
```
当up主发送动态后，会推送到设置的平台，如果安装了ffmpeg，可以自动备份新稿件到alist
![](https://example.com/undefined20250307003814757.png)
## 直播间监控
```json5
{
  "Tracing": ["544853"], //要监控的直播间，类型为string[]
}

```
主播开播后会发送推送,同时记录直播间内弹幕和礼物。


![](https://example.com/undefined20250307003957739.png)
![](https://example.com/undefined20250310000627967.png)
<details>
 <summary> 展开截图 </summary>
<img src="https://example.com/undefined20250307004410454.png">
<img src="https://example.com/undefined20250307004703580.png"> 
<img src="https://example.com/undefined20250307005721329.png">
<img src="https://example.com/undefined20250510135208967.png">
</details>

## 视频下载
下载视频到Alist，需要安装ffmpeg。 入口在网页端顶部下拉菜单。
支持下载分P视频，单独的视频，播放列表
<img src="https://example.com/undefined20250307010130538.png">
<img src="https://example.com/undefined20250307010229305.png">

## 分区监控

每隔一定时间监控分区内所有主播，并记录大航海和粉丝数据，可在前端页面查看趋势。

主播在分区内直播时：数据在两天内更新过则不更新。

程序最多同时连接20个直播间ws服务器，如果当前不到20个，会选择从分区内选择直播间连接。

支持设置http代理服务器。

支持分布式

~~~jsonc
{
  "Mode": "Master",
  "Slaves": ["http://127.0.0.1:30001","http://127.0.0.1:30002","http://127.0.0.1:30003"],
 }
~~~
设置后会自动从分区获取正在直播的直播间并将任务下发到子节点（包括主节点自身


主节点只负责下发任务（可以下发给自己），所以需要确保子节点也能访问到数据库。

若不需要分布式爬虫，Slaves留空即可

## 观众分析

![](https://example.com//img/20250728231920.png)
![](https://example.com//img/20250728232125.png)
# Development

- 运行main.test.go的TestHttp方法可以启动一个测试环境，不会爬取直播间，用于测试和调试接口
- main.go的main方法，启动正常的环境
- 前端页面在Page目录下，运行vite启动 


# TODO
- [ ] 完善文档
- [ ] 程序启动时检查配置文件正确性
# Referer
接口文档：[bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect)

wbi算法：[CuteReimu/bilibili/](https://github.com/CuteReimu/bilibili/)

