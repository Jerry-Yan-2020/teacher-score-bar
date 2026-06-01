# 课堂加减分悬浮条

一个 Windows 桌面小程序，用于老师在播放 PPT 时悬浮显示学生名单，并快速给学生加分或减分。

## 使用方式

1. 运行 `dist\TeacherScoreBar.exe`。
2. 屏幕中上方会出现悬浮评分条。
3. 鼠标或触屏在学生区域左右拖动可以滑动名单。
4. 点击学生姓名，选择加分或减分。
5. 点击悬浮条左侧的“后台”打开管理窗口，可配置班级、学生姓名和总分。
6. 后台右侧可以设置悬浮窗颜色、学生区域颜色、学生卡颜色、加分色、减分色和透明度。
7. 选择学生后，后台会显示该学生每一笔加减分记录，包括时间、变化值和变化前后总分。

数据保存在当前 Windows 用户的配置目录：

`%APPDATA%\TeacherScoreBar\data.json`

运行日志保存在：

`%APPDATA%\TeacherScoreBar\scorebar.log`

如果程序异常退出或无响应，请把这个日志文件发给开发者排查。

## GitHub 发布

GitHub 不支持使用账号密码直接推送代码。请使用 GitHub CLI 登录，或创建 Personal Access Token 后设置为远程仓库认证。

## 开发构建

```powershell
go build -ldflags="-H windowsgui" -o dist\TeacherScoreBar.exe .
```
