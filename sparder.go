package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sparder/spider"
	"strings"
	"time"
)

var (
	versionFlag *string = flag.String("v", "0.9.1", "Print the version number.")
	historyUrl  *string = flag.String("url", "", "The spark history url.")
	dir         *string = flag.String("dir", filepath.Dir(os.Args[0]), "Save the markdown file.")
	tp          *string = flag.String("type", "sp", "Save type: ps(PrintSimple)/pp(PrintPage)/ss(SaveSimple)/sp(SavePage).")
	pages               = []string{"jobs", "stages", "storage", "environment", "executors"}

	hurl string
	page string
	eurl string
)

func main() {
	flag.Parse()
	
	test := "http://vrv207:18080/history/application_1542875136548_0630/appattempt_1542875136548_0630_000002/executors/"
	historyUrl = &test
	hurl = getUrl(*historyUrl)
	fmt.Printf("**Track URL:** %s\n", hurl)
	if "" == hurl {
		return
	}
	
	jobs()
	stages()
	storage()
	executors()
	environment()
	
	switch *tp {
	case "ps", "PrintSimple":
		spider.PrintSimple()
	case "pp", "PrintPage":
		spider.PrintPage()
	case "ss", "SaveSimple":
		file := createFile()
		defer file.Close()
		spider.SaveSimple(file)
	case "sp", "SavePage":
		file := createFile()
		defer file.Close()
		spider.SavePage(file)
	}
}

func createFile() *os.File {
	path := filepath.Join(*dir, "track_"+time.Now().Format("20060102150405")+".md")
	fmt.Printf("create markdown file: %s\n", path)
	file, _ := os.Create(path)
	fmt.Fprintf(file, "**Track URL:** %s\n", hurl)
	return file
}

func jobs() {
	eurl = hurl + "/jobs"
	page = spider.GetPage(eurl)
	jobs := spider.BuildJobs(page)
	spider.AddPage("Jobs", eurl, jobs)
}

func stages() {
	eurl = hurl + "/stages"
	page = spider.GetPage(eurl)
	stages := spider.BuildStages(page)
	spider.AddPage("Stages", eurl, stages)
}

func storage() {
	eurl = hurl + "/storage"
	page = spider.GetPage(eurl)
	storage := spider.BuildStorage(page)
	spider.AddPage("Storage", eurl, storage)
}

func executors() {
	eurl = hurl + "/executors"
	page = spider.GetPage(eurl)
	executors := spider.BuildExecutors(page)
	spider.AddPage("Executors", eurl, executors)
}

func environment() {
	eurl = hurl + "/environment"
	page = spider.GetPage(eurl)
	environment := spider.BuildEnvironment(page)
	spider.AddPage("Environment", eurl, environment)
}

func getUrl(s string) string {
	for _, p := range pages {
		i := strings.LastIndex(s, "/"+p)
		if i != -1 {
			return s[:i]
		}
	}
	return s
}
