package sparder

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	//	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	savePath  = filepath.Join(filepath.Dir(os.Args[0]), "sparder_"+time.Now().Format("200601021504"))
	everyPage = []string{"jobs", "stages", "storage", "environment", "executors"}
)

func SaveHtml(baseUrl string) {
	for _, p := range everyPage {
		savePage16(baseUrl+"/"+p, p)
	}
}

/**
*spark 1.6 history
 */
func savePage16(url string, pageName string) {
	_, err := os.Stat(savePath)
	if nil != err {
		fmt.Printf("create %s\n", savePath)
		os.MkdirAll(savePath, 755)
	}
	htmlFilePath := filepath.Join(savePath, pageName+".html")
	download(url, htmlFilePath)
	resource := saveResource(url, htmlFilePath)
	redirectSource(resource, htmlFilePath)
}

func saveResource(url string, filePath string) []string {
	file, _ := os.Open(filePath)
	defer file.Close()
	sc := bufio.NewScanner(file)

	r, _ := regexp.Compile("(href|src)=\"([^\"]*)\"")
	resource := make([]string, 0)

	resource = append(resource, "/static/spark-logo-77x50px-hd.png")

	for sc.Scan() {
		line := sc.Text()
		tmp := r.FindAllStringSubmatch(line, -1)
		for _, tmp2 := range tmp {
			if strings.Contains(tmp2[2], "static") {
				resource = append(resource, tmp2[2])
				continue
			}
			if !strings.HasSuffix(tmp2[2], "/") && strings.Contains(tmp2[2], "history") {
				notPage := true
				for _, p := range everyPage {
					if strings.HasSuffix(tmp2[2], p) {
						notPage = false
						fmt.Println(1111)
						break
					}
				}
				if notPage {
					resource = append(resource, tmp2[2])
				}
			}
		}
	}

	baseUrl := url[:strings.Index(url[7:], "/")+7]

	for _, s := range resource {
		localPath := savePath + strings.Replace(s, "/", string(filepath.Separator), -1)
		url := baseUrl + s
		if strings.Contains(localPath, "?") {
			localPath = strings.Replace(localPath, "?", "-", -1)
			localPath = strings.Replace(localPath, "=", "-", -1)
			localPath = strings.Replace(localPath, "&amp;", "-", -1)
			localPath = localPath + ".html"
			url = strings.Replace(url, "&amp;", "&", -1)
		}
		download(url, localPath)
		if strings.HasSuffix(localPath, ".html") {
			redirectSource(resource, localPath)
		}
	}

	return resource
}

func download(url string, local string) {
	os.MkdirAll(filepath.Dir(local), 755)
	file, _ := os.Create(local)
	defer file.Close()
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	io.Copy(file, resp.Body)
}

func redirectSource(resource []string, filePath string) {
	file, _ := os.Open(filePath)
	defer file.Close()
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	sc := bufio.NewScanner(file)
	headStart, headEnd := false, false
	logo := false
	navStart, navEnd := false, false
	for sc.Scan() {
		line := sc.Text()
		line, headStart, headEnd = fixHead(line, resource, headStart, headEnd)
		line, logo = fixLogo(line, logo)
		line, navStart, navEnd = fixNav(line, navStart, navEnd)
		line = fixHistory(line)
		buff.WriteString(line)
		buff.WriteRune('\n')
	}
	file2, _ := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 755)
	defer file2.Close()
	file2.Write(buff.Bytes())
}

func fixHead(line string, resource []string, start bool, end bool) (string, bool, bool) {
	if end {
		return line, start, end
	}
	l := line
	s := start
	e := end
	if strings.Contains(l, "<head>") {
		s = true
	}
	if s {
		for _, r := range resource {
			l = strings.Replace(l, r, "."+r, -1)
		}
	}
	if strings.Contains(l, "</head>") {
		e = true
	}
	return l, s, e
}

func fixLogo(line string, logo bool) (string, bool) {
	if logo {
		return line, logo
	}
	if strings.Contains(line, "/static/spark-logo-77x50px-hd.png") {
		return strings.Replace(line, "/static/spark", "./static/spark", 1), true
	}
	return line, logo
}

func fixNav(line string, start bool, end bool) (string, bool, bool) {
	if end {
		return line, start, end
	}
	l := line
	s := start
	e := end
	if strings.Contains(l, "class=\"nav\">") {
		s = true
	}
	if s {
		idx := strings.Index(l, "href=")
		if -1 != idx {
			for _, p := range everyPage {
				idx2 := strings.Index(l, p)
				if -1 != idx2 {
					l = l[:idx] + "href=\"" + p + ".html" + l[idx2+len(p)+1:]
					break
				}
			}
		}
	}
	if strings.Contains(l, "</ul>") {
		e = true
	}
	return l, s, e
}

func fixHistory(line string) string {
	if strings.Contains(line, "history") && strings.Contains(line, "?") {
		r, _ := regexp.Compile("href=\"([^\"]*)")
		old := r.FindStringSubmatch(line)
		if len(old) < 1 {
			return line
		}
		l := old[1]
		l = strings.Replace(l, "?", "-", -1)
		l = strings.Replace(l, "=", "-", -1)
		l = strings.Replace(l, "&amp;", "-", -1)
		l = "." + l + ".html"
		return strings.Replace(line, old[1], l, 1)
	}
	return line
}
