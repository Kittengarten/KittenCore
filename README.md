# KittenCore

基于 [ZeroBot](https://github.com/wdvxdr1123/ZeroBot) 开发的 QQ 机器人程序

## 使用方法

以 `Bot` 的昵称为开头的发言会被视为与 `Bot` 对话。

由于项目中可能含有非 `ASCII` 字符文件名，可能需要执行 `git config --bool hooks.allownoascii true`。

由于项目依赖了使用 `go:linkname` 的包，必须在编译命令中添加 `-ldflags=-checklinkname=0` 参数，可直接使用本项目的 `makefile` 编译（已经添加了该参数）。

包含的插件可用于任何基于 [ZeroBot](https://github.com/wdvxdr1123/ZeroBot) 的程序，但在一些环境中可能需要部分修改。部分功能需要通过其它程序实现。

本程序也可以 `import` 的方式添加任何 [ZeroBot](https://github.com/wdvxdr1123/ZeroBot) 的插件。
