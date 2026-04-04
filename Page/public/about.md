<div style="display: flex; flex-direction: column">
<div style="display: flex ;justify-content: center">
<img src="/telegram.svg" style="width:24px;height:24px"></img>
<a classname="ml-2" href="https://t.me/aicu_buzz" target="_blank">Telegram群组(AICU聊天群)</a>
</div>
<div style="display: flex ;justify-content: center" class="mt-2">
<img src="/mail.svg" style="width:24px;height:24px"></img>
<a classname="ml-2" href="mailto:admin@shiohix.resend.app" target="_blank">邮箱</a>
</div>
<div style="display: flex ;justify-content: center" class="mt-2">
<a classname="ml-2" href="https://vtb.fider.io/" target="_blank">bug反馈/追踪</a>
</div>
</div>



把每页的默认大小改成了100，如果不习惯可以在Misc -> 设置中修改。


关注<a href="https://space.bilibili.com/3546757543758795" target="_blank">铃见Suzumi</a>谢谢喵

<bili-dynamic-card OID="1112083250515279896"/>




## Changelog

### 26.2.28

- 新的首页


### 26.2.26

- 饼状图显示问题应该修好了。
- 修复用户的大航海图标显示错误

### 25.12.8

* 在桌面端，主播详情页的直播卡片以每行4列显示
* 你可以在设置页面 设置每页默认数据条数

### 25.11.23

* 搜索页面显示当前热门直播间

### 25.11.21

* 更新这个文档

### 25.11.14

* 添加留言功能
* 添加历史动态查询
* 优化用户成分图表显示
* 优化移动端表格显示

### 25.11.11

* 添加大航海水分查询
* 修复部分情况下分页页数错误返回

### 25.10.5
* 修复前端页面部分元素没有对齐
* 添加打米排行榜

### 25.9.30
* 修复可能出现的panic
* status接口更换成websocket
* 多线程刷新用户粉丝
* 优化刷新分区的逻辑，避免一次刷新分区时间过久导致下一次刷新需要很长时间
* 添加同接记录
* 调整表结构

### 25.9.18
* 修复部分对话框滚动问题
* 修复直播结束时间没有正确保存
* 支持按主播公会筛选


### 25.9.12
* 清理代码
* 优化内存占用
* 修复动态解析的bug
* 直播详情界面可以显示当前直播流

### 25.9.11
* 使用recharts的图表
* 更换LAPLACE提供的头像api
* 清理代码


### 25.9.4
* 异步获取大航海数据，避免高峰期过长时间地阻塞协程
* 空闲时期降低被爬取主播的粉丝要求，5000粉丝即可被记录。

### 25.8.30
* 修复月增排行的一些错误
* 添加这个更新日志的对话框
* 更新React到19.1.1
* 使用React Compiler

### 25.8.27
- 白嫖b站的cdn，前端加载速度大幅提高
- 后端api更换到香港的节点，三网直连速度优异，如加载较慢请关闭代理或者将本站设为直连

### 25.8.13
- 支持查看主播的历史大航海成员
- 支持对两个时间的大航海成员比较差异