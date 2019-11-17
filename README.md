# m3u8视频下载

根据m3u8文件并发下载视频，速度快。

支持系统：mac，linux

m3u8_download是已经编译好的二进制文件,packages没有弄

如果是go语言使用的同学可以直接下载源码运行，里面有福利奥：
```shell
$ go run main.go -out 我在jy的日子.mp4
$ ./m3u8_downloader -out 我在jy的日子.mp4
```
下载m3u8视频的话
```shell
-out string
    -out A.mp4
-url string
    -url https://www.00h.tv/index.m3u8 (default "https://youku.com-zx-iqiyi.com/20190918/7221_5a8e267f/index.m3u8")
```

