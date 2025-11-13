<div style="display: flex; flex-direction: column">
<div style="display: flex ;justify-content: center">
<img src="/telegram.svg" style="width:24px;height:24px"></img>
<a classname="ml-2" href="https://t.me/aicu_buzz" target="_blank">Telegram群组</a>
</div>
<div style="display: flex ;justify-content: center" class="mt-2">
<img src="/github.svg" style="width:24px;height:24px"></img>
<a classname="ml-2" href="https://github.com/114514ns/BiliMonitor" target="_blank">github</a>
</div>
<div style="display: flex ;justify-content: center" class="mt-2">
<a classname="ml-2" href="mailto:admin@ikun.dev" target="_blank">邮箱</a>
</div>
</div>

## Notice
- 不建议公开宣传本站
- 有任何建议，问题，想要的功能，欢迎反馈

## Changelog

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
* 
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