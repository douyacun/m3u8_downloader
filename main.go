package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// 最大并发数
const MAXREQ = 10

// m3u8层级
var m3u8Index int

// 下载目录
var tempDir string

// 限制并发数量
var sem = make(chan int, MAXREQ)

func main() {
	var (
		s   string
		out string
	)
	flag.StringVar(&s, "url", "https://youku.com-zx-iqiyi.com/20190918/7221_5a8e267f/index.m3u8", "-url https://www.00h.tv/index.m3u8")
	flag.StringVar(&out, "out", "", "-out A.mp4")
	flag.Parse()
	if out == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	// 初始化存储目录
	md5ctx := md5.New()
	md5ctx.Write([]byte(s))
	tempDir = fmt.Sprintf("%s/%s/%s", os.Getenv("HOME"), "Downloads", hex.EncodeToString(md5ctx.Sum(nil)))
	if !FileExist(tempDir) {
		if err := os.Mkdir(tempDir, 0744); err != nil {
			log.Fatalf("%s 创建失败", tempDir)
		}
	}
	download([]string{s})
	ffmpeg(out)
}

func ffmpeg(out string) {
	downloadsDir := fmt.Sprintf("%s/Downloads", os.Getenv("HOME"))
	command := fmt.Sprintf("cd %s && ffmpeg -i %d.m3u8 -c copy %s/%s", tempDir, m3u8Index, downloadsDir, out)
	cmd := exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		log.Fatalf("%s\n执行失败, %s", command, err)
	}
	log.Printf("文件下载成功: %s/%s", downloadsDir, out)
}

func getM3u8Index() int {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	m3u8Index++
	return m3u8Index
}

func download(list []string) {
	wg := sync.WaitGroup{}
	var (
		err error
		res []byte
	)
	// 下载目录
	for _, v := range list {
		prefix := parsePrefix(v)
		if strings.Contains(v, ".m3u8") {
			m3u8F := fmt.Sprintf("%s/%d.m3u8", tempDir, getM3u8Index())
			if !FileExist(m3u8F) {
				res, err = httpGet(v)
				if err != nil {
					log.Fatalf("get %s failed, %s", v, res)
				}
				if err = ioutil.WriteFile(m3u8F, res, 0666); err != nil {
					log.Fatalf("%s 写入失败, %s", m3u8F, err)
				}
			} else {
				res, err = ioutil.ReadFile(m3u8F)
				if err != nil {
					log.Fatalf("%s 读取失败, %s", m3u8F, err)
				}
			}
			log.Printf("m3u8文件(%s)下载成功", v)
			download(parseM3u8(res, prefix))
		} else if strings.Contains(v, ".ts") {
			wg.Add(1)
			sem <- 1
			go downloadTsF(v, &wg)
		} else {
			log.Fatalf("未知文件类型, %s", v)
		}
	}
	wg.Wait()
}

func downloadTsF(u string, wg *sync.WaitGroup) {
	defer wg.Done()
	f := parseFile(u)
	tsF := fmt.Sprintf("%s/%s", tempDir, f)
	if !FileExist(tsF) {
		res, err := httpGet(u)
		if err != nil {
			log.Printf("ts文件(%s)下载失败", err)
			return
		}
		if err = ioutil.WriteFile(tsF, res, 0644); err != nil {
			log.Printf("ts文件(%s)写入失败", tsF)
			return
		}
		log.Printf("%s 写入成功", tsF)
	} else {
		log.Printf("%s 文件已经存在", tsF)
	}
	<-sem
	return
}

func parsePrefix(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		log.Fatalf("url parse failed: %s", err)
	}
	dir := u.Path[:strings.LastIndex(u.Path, "/")]
	return fmt.Sprintf("%s://%s%s/", u.Scheme, u.Host, dir)
}

func parseFile(s string) string {
	return s[strings.LastIndex(s, "/")+1:]
}

func parseM3u8(content []byte, prefix string) []string {
	var list = make([]string, 0, strings.Count(string(content), "\n"))
	for _, v := range strings.Split(string(content), "\n") {
		v = strings.TrimSpace(v)
		if !strings.HasPrefix(v, "#") && len(v) > 0 {
			list = append(list, prefix+v)
		}
	}
	return list
}

func FileExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Fatal(err)
		}
	}
	return true
}

func httpGet(u string) ([]byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
