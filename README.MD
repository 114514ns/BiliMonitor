# BiliMonitor
~~视奸你喜欢的up主~~
# Features
- 定时记录关注列表内的up主的粉丝数
- 记录直播间内弹幕，礼物，营收
- 当主播开始直播/发布动态时发送邮件/Server酱/QQ提醒
- 记录主播大航海
- 有一个后台web，可以在上面查看记录到的数据
- 备份新稿件到Alist
- 联动Mikufans录播姬，可选单独备份直播音频，转码为mp4
- 可选MySQL和SQLite数据库
- 备份(爬取)用户动态和评论

# Screenshots

![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com/undefinedqq_pic_merged_1740929772073.jpg)
![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com/undefined20250305001302952.png)
![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com/undefined20250305001138294.png)
![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com/undefined20250302233820130.png)
![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com//img/20250212235435.png)
![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com/undefined20250220222834082.png)
![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com/undefined20250302233902605.png)
![](https://imgbed-1254007525.cos.ap-nanjing.myqcloud.com/undefined20250302233924636.png)

# Configuration

```json5
{
  "RefreshFollowingsDelay": "30m", //刷新关注列表间隔
  "User": "451537183", //你的uid，列表内的所有up主会被定时记录粉丝量，需要设置关注列表允许公开访问
  "CommonDelay": "30m", //记录粉丝量的间隔
  "SpecialList": [], //特别关注列表，列表内的up主发布的动态和稿件会被推送
  "SpecialDelay": "2m", //刷新特别关注列表内up主动态的间隔
  "Cookie": "", //b站cookie，f12随便找个请求然后复制cookie粘贴过来
  "EnableEmail": true, //启用邮件推送，目前使用resend api实现发送邮件
  "ResendToken": "re_TLeNcEDu_Ht8QFPBRPH6JyKZjfnxmztwB", //申请的token,一天免费100封邮件，你可以暂时用我的来测试。
  "FromMail": "bili@ikun.dev", //发件邮箱
  "ToMail": [], //收件邮箱
  "EnableQQBot": false, //启用OneBot推送
  "ReportTo": [
    "3212329718" //要推送到的qq账号
  ],
  "BackServer": "http://127.0.0.1:3090", //OneBot后端
  "EnableServerPush": true, //启用Server酱推送，参考https://sc3.ft07.com/接入
  "ServerPushKey": "", //Server酱key
  "Tracing": [], //要监控的直播间，类型为string[]
  "EnableAlist": true, //启用alist相关功能
  "AlistServer": "http://192.168.10.16:5244/", //alist服务器，注意url最后的斜杠
  "AlistUser": "admin", //alist用户名
  "AlistPass": "Zyh060813@", //alist密码
  "AlistPath": "移动/Bili/", //要保存到的路径
  "EnableLiveBackup": true, //启用直播备份
  "MikuPath": "/mnt/c/Users/zhang/Desktop/Record/", //录播姬路径
  "EnableSQLite": false, //启用SQLite数据库
  "SQLitePath": "database.db", //sqlite数据库名字
  "EnableMySQL": true, //启用MySQL数据库
  "SQLName": "bili", //数据库名
  "SQLUser": "bili", //用户名
  "SQLPass": "CERtxrAH2zaik7b2", //密码
  "SQLServer": "127.0.0.1:3306", //数据库地址
  "CodeToMP4": true, //将录播姬的flv格式转为mp4格式
  "SplitAudio": true //上传录播时另外上传一份直播的声音
}
```
# TODO
- [ ] 自动刷新Cookie
- [ ] 完善文档
- [ ] 备份动态的评论？

# Referer
[bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect)
