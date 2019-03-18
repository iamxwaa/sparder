package local

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	home                = getHome()
	alreadyCrawl        = map[string]int{}
	alreadyDownload     = map[string]int{}
	alreadySetLocalPath = map[string]int{}
)

type FileInfo struct {
	localDir  string
	localFile string
	name      string
	suffix    string
	isPage    bool
}

type ResourceInfo struct {
	fi  FileInfo
	url string
}

type PageInfo struct {
	fi     FileInfo
	url    string
	domain string
	rs     []ResourceInfo
}

func crawlPage(curl string) {
	url := fixUrl(curl)
	//检查页面是否爬取过
	if checkCrawled(url) {
		fmt.Printf("skip crawl %s\n", url)
		return
	}
	fmt.Printf("crawl %s\n", url)
	//获取页面信息
	pi := getPageInfo(url)
	//下载页面到本地
	download(pi.url, pi.fi)
	//获取页面中所引用到的资源
	pi.rs = getResourceFromPage(pi)
	//下载页面中引用的资源
	for _, r := range pi.rs {
		url2 := r.url
		if !strings.HasPrefix(url2, "http://") {
			url2 = pi.domain + url2
		}
		if r.fi.isPage {
			crawlPage(url2)
		} else {
			download(url2, r.fi)
		}
	}
	//将页面中引用的资源路径改为本地相对路径
	setResourceLocalPath(pi)
}

func getPageInfo(url string) PageInfo {
	pi := PageInfo{}
	pi.domain = getDomain(url)
	pi.fi.name = getPageName(url)
	pi.url = url
	pi.fi.suffix = getSuffix(url)
	pi.fi.isPage = (pi.fi.suffix == ".html")
	setLocal(&pi.fi, pi.url, len(pi.domain), len(pi.url)-len(pi.fi.name)-1)
	return pi
}

func getPageName(url string) string {
	u := url
	if strings.HasSuffix(u, "/") {
		u = u[:len(u)-1]
	}
	i := strings.LastIndex(u, "/") + 1
	u = u[i:]
	u = strings.Replace(u, "?", "-", -1)
	u = strings.Replace(u, "=", "-", -1)
	u = strings.Replace(u, "&amp;", "-", -1)
	u = strings.Replace(u, "&", "-", -1)
	return u
}

func fixUrl(url string) string {
	u := url
	u = strings.Replace(u, "&amp;", "&", -1)
	return u
}

func getDomain(url string) string {
	i := strings.Index(url[7:], "/")
	return url[:7+i]
}

func getSuffix(url string) string {
	if strings.HasSuffix(url, ".js") {
		return ".js"
	}
	if strings.HasSuffix(url, ".css") {
		return ".css"
	}
	if strings.HasSuffix(url, ".jpg") {
		return ".jpg"
	}
	if strings.HasSuffix(url, ".png") {
		return ".png"
	}
	if strings.HasSuffix(url, ".gif") {
		return ".gif"
	}
	return ".html"
}

func setLocal(fi *FileInfo, url string, subs int, sube int) {
	fi.localDir = filepath.Join(home, strings.Replace(url[subs:sube], "/", string(filepath.Separator), -1))
	if fi.isPage {
		fi.localFile = filepath.Join(fi.localDir, fi.name+fi.suffix)
	} else {
		fi.localFile = filepath.Join(fi.localDir, fi.name)
	}
}

func checkCrawled(url string) bool {
	if alreadyCrawl[url] != 0 {
		return true
	}
	alreadyCrawl[url] = 1
	return false
}

func checkDownload(url string) bool {
	if alreadyDownload[url] != 0 {
		return true
	}
	alreadyDownload[url] = 1
	return false
}

func checkSetLocalPath(local string) bool {
	if alreadySetLocalPath[local] != 0 {
		return true
	}
	alreadySetLocalPath[local] = 1
	return false
}

func download(url string, fi FileInfo) {
	if checkDownload(url) {
		fmt.Printf("skip download %s\n", url)
		return
	}
	os.MkdirAll(fi.localDir, 755)
	file, _ := os.Create(fi.localFile)
	defer file.Close()
	fmt.Printf("download %s to %s\n", url, fi.localFile)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	io.Copy(file, resp.Body)
}

func getResourceFromPage(pi PageInfo) []ResourceInfo {
	reader, _ := os.Open(pi.fi.localFile)
	defer reader.Close()
	sc := bufio.NewScanner(reader)
	r, _ := regexp.Compile("(href|src)=\"([^\"]+)\"")
	rs := make([]ResourceInfo, 0)
	var filter map[string]int = map[string]int{}
	for sc.Scan() {
		line := sc.Text()
		links := r.FindAllStringSubmatch(line, -1)
		if len(links) != 0 {
			for _, tlks := range links {
				testlink := tlks[2]
				if filter[testlink] == 0 && !strings.HasPrefix(testlink, "#") && len(testlink) > 1 {
					filter[testlink] = 1
					ri := ResourceInfo{}
					ri.url = testlink
					ri.fi.suffix = getSuffix(testlink)
					ri.fi.isPage = (ri.fi.suffix == ".html")
					ri.fi.name = getPageName(testlink)
					setLocal(&ri.fi, ri.url, 0, len(ri.url)-len(ri.fi.name)-1)
					rs = append(rs, ri)
				}
			}
		}
	}
	return rs
}

func setResourceLocalPath(pi PageInfo) {
	if !pi.fi.isPage && checkSetLocalPath(pi.fi.localFile) {
		fmt.Printf("skip set resource local path %s\n", pi.fi.localFile)
		return
	}
	dot := strings.Count(pi.fi.localDir[len(home):], string(filepath.Separator))
	//获取页面中的资源路径
	reader, _ := os.Open(pi.fi.localFile)
	defer reader.Close()
	sc := bufio.NewScanner(reader)
	r, _ := regexp.Compile("(href|src)=\"([^\"]+)\"")
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	for sc.Scan() {
		line := sc.Text()
		links := r.FindAllStringSubmatch(line, -1)
		if len(links) != 0 {
			for _, tlks := range links {
				testlink := tlks[2]
				if !strings.HasPrefix(testlink, "#") && len(testlink) > 1 {
					rlink := getRelativePath(testlink, dot)
					line = strings.Replace(line, testlink, rlink, -1)
				}
			}
		}
		buff.WriteString(line)
		buff.WriteRune('\n')
	}

	file2, _ := os.OpenFile(pi.fi.localFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 755)
	defer file2.Close()
	file2.Write(buff.Bytes())
}

func getRelativePath(url string, dot int) string {
	u := url
	if strings.HasSuffix(u, "/") {
		u = u[:len(u)-1]
	}
	if strings.HasPrefix(u, "http://") {
		u = u[7:]
		u = u[strings.Index(u, "/"):]
	}
	isPage := getSuffix(u) == ".html"
	if isPage {
		name := getPageName(u)
		u = u[:strings.LastIndex(u, "/")]
		u = u + "/" + name + ".html"
	}
	d := ""
	for i := 0; i < dot; i++ {
		d += "../"
	}
	return d[:len(d)-1] + u
}

func getHome() string {
	base := filepath.Join(filepath.Dir(os.Args[0]), "sparder")
	for i := 1; i < 999; i++ {
		num := ""
		if i < 10 {
			num = "_00" + strconv.Itoa(i)
		}
		if i < 100 && i > 9 {
			num = "_0" + strconv.Itoa(i)
		}
		if i < 1000 && i > 99 {
			num = "_" + strconv.Itoa(i)
		}
		_, err := os.Stat(base + num)
		if nil != err {
			return base + num
		}
	}
	return base
}
