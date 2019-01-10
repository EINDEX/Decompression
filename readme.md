# Decomperssion 解压工具

工具运行时 讲会把指定目录下的文件移动或者删除,
如果源文件没有备份,请勿使用本工具!

本工具不支持 32位系统

Usage of ./decomperssion:
  -w string
    	工作路径 默认 worker (default "worker")

## 使用方法

### Windows X64
相同目录下 有 `7z.exe`  `7z.dll` `7-zip.dll`
decomperssion.exe -w=<目录名>

### MacOS
`brew install p7zip`
`./decomperssion_macos -w=<目录名>`

### Linux X64
`apt install p7zip`
`./decomperssion_linux_x64 -w=<目录名>`