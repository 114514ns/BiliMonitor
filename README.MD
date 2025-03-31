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

# Screenshots

![](https://example.com/undefined20250305001302952.png)
![](https://example.com/undefined20250305001138294.png)
![](https://example.com/undefined20250302233820130.png)
![](https://example.com//img/20250212235435.png)
![](https://example.com/undefined20250220222834082.png)
![](https://example.com/undefined20250302233902605.png)
![](https://example.com/undefined20250302233924636.png)

# Known issues
- 偶有直播开始或结束事件无法正确检测
# Configuration

```json5
{
  "RefreshFollowingsDelay": "30m", //刷新关注列表间隔
  "User": "2", //你的uid，列表内的所有up主会被定时记录粉丝量，需要设置关注列表允许公开访问
  "CommonDelay": "30m", //记录粉丝量的间隔
  "SpecialList": [], //特别关注列表，列表内的up主发布的动态和稿件会被推送，类型为int
  "SpecialDelay": "2m", //刷新特别关注列表内up主动态的间隔
  "Cookie": "", //b站cookie，f12随便找个请求然后复制cookie粘贴过来
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
  "Tracing": [], //要监控的直播间，类型为string[]
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
  "SplitAudio": true //上传录播时另外上传一份直播的声音
}
```
大部分功能都需要Cookie，所以请先填写好Cookie。强烈建议使用小号，虽然不一定会有啥问题，但是还是不要用大号。
<br>
不是所有功能都需要数据库，但是还是请先启用数据库功能。
## 粉丝量监控
```json5
"User": "2", //你的uid，列表内的所有up主会被定时记录粉丝量，需要设置关注列表允许公开访问
```
这里改成你的uid，然后允许公开访问你的关注列表即可。
<br/
目前没有前端界面显示粉丝量的趋势，只存储在数据库里，后面会加。

## 动态监控
```json5
  "SpecialList": [12434430], //特别关注列表，列表内的up主发布的动态和稿件会被推送，类型为int
  "SpecialDelay": "2m", //刷新特别关注列表内up主动态的间隔
```
当up主发送动态后，会推送到设置的平台，如果安装了ffmpeg，可以自动备份新稿件到alist
![](https://example.com/undefined20250307003814757.png)
## 直播间监控
```json5
"Tracing": [], //要监控的直播间，类型为string[]
```
主播开播后会发送推送,同时记录直播间内弹幕和礼物。
可以在web端查看，端口8080
![](https://example.com/undefined20250307003957739.png)
![](https://example.com/undefined20250310000627967.png)
<details>
 <summary> 展开截图 </summary>
<img src="https://example.com/undefined20250307004410454.png">
<img src="https://example.com/undefined20250307004703580.png"> 
<img src="https://example.com/undefined20250307005721329.png">
</details>

## 视频下载
下载视频到Alist，需要安装ffmpeg。 入口在网页端顶部下拉菜单。
支持下载分P视频，单独的视频，播放列表
<img src="https://example.com/undefined20250307010130538.png">
<img src="https://example.com/undefined20250307010229305.png">
# TODO
- [ ] 自动刷新Cookie
- [ ] 完善文档
- [ ] 备份动态的评论？
- [ ] 录播时将视频为HEVC/AV1以减少文件体积
- [ ] 程序启动时检查配置文件正确性

# Referer
[bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect)
<br/>
[CuteReimu/bilibili/](https://github.com/CuteReimu/bilibili/)

