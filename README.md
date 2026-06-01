# 课堂加减分悬浮条

一个 Windows 桌面小程序，用于老师在播放 PPT 或其他教学软件时，在屏幕上方悬浮显示学生名单，并快速给学生加分、减分、查看总分和历史记录。

## 功能

- 悬浮评分条置顶显示，不影响 PPT 正常播放和操作。
- 悬浮评分条支持拖拽移动，并可从边缘或角落拖拽缩放。
- 学生名单支持鼠标或触屏左右滑动。
- 点击学生姓名后可快速加分、减分或清零。
- 后台可配置班级、学生姓名和学生总分。
- 后台可查看每个学生的每一笔加减分记录，包括时间、变化值、变化前后总分和来源。
- 支持设置悬浮窗、学生区域、学生卡片、加分色、减分色、强调色和透明度。
- 使用 Windows 原生取色器选择颜色。
- 数据保存到当前 Windows 用户目录，关闭程序后不会丢失。

## 下载与使用

如果只是使用程序，不需要安装 Go，也不需要安装任何开发环境。

1. 下载或复制 `dist\TeacherScoreBar.exe` 到 Windows 电脑。
2. 双击运行 `TeacherScoreBar.exe`。
3. 屏幕中上方会出现悬浮评分条。
4. 点击悬浮条左侧的“后台”按钮，配置班级和学生。
5. 在悬浮条中点击学生姓名，选择加分或减分。

如果要发给其他老师，直接发送下面这个文件即可：

`dist\TeacherScoreBar.exe`

建议也可以把它压缩成 zip 后发送，避免部分聊天软件拦截 exe 文件。

## 数据与日志

学生、班级、分数、样式和加减分记录保存在当前 Windows 用户配置目录：

`%APPDATA%\TeacherScoreBar\data.json`

运行日志保存在：

`%APPDATA%\TeacherScoreBar\scorebar.log`

如果程序异常退出或无响应，可以查看或反馈这个日志文件。

## 开发环境

- Windows
- Go 1.26 或更高版本

## 构建

在仓库根目录执行：

```powershell
go build -ldflags="-H windowsgui" -o dist\TeacherScoreBar.exe .
```

生成后的 `dist\TeacherScoreBar.exe` 是普通 Windows 可执行文件，其他 Windows 电脑不需要 Go 环境即可运行。

## 开源协议

本项目使用 MIT License，详见 [LICENSE](LICENSE)。
