package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetPageToLocal(url string) {
	crawlPage(url)
	makeIndex()
}

func makeIndex() {
	file, _ := os.Create(filepath.Join(home, "index.html"))
	defer file.Close()
	file.WriteString(getIndexHtml())
}

func getIndexSrc() string {
	index := ""
	filepath.Walk(home, func(path string, info os.FileInfo, err error) error {
		if "" != index {
			return nil
		}
		if info.Name() == "jobs.html" {
			index = "." + strings.Replace(path[len(home):], string(filepath.Separator), "/", -1)
		}
		return nil
	})
	fmt.Printf("set home page to %s\n", index)
	return index
}

func getIndexHtml() string {
	return `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-type" content="text/html; charset=utf-8"/>        
    <title>Spark Jobs</title>
    <style>
        body {margin: 0px;padding: 0px;overflow: hidden;}
        iframe {border: none;width: 100%;height: calc(100% - 32px);}
        #footer {
		    height: 32px;
		    text-align: center;
		    font-size: 12px;
		    color: #999999;
		    background-image: linear-gradient(to bottom, #e5e5e5, #ffffff);
		    box-shadow: 0px -1px 10px rgba(0,0,0,.1);
		    width: 100%;
		    border: 1px solid #d4d4d4;
        }
    </style>
  </head>
  <body>
    <iframe src="` + getIndexSrc() + `"></iframe>
  </body>
  <div id="footer">Create By sparder@XW</div>
  <script type="text/javascript">
    document.body.style.height=(screen.availHeight-100)+"px"
  </script>
</html>`
}
