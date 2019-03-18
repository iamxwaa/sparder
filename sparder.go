package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sparder/sparder"
	"sparder/sparder/local"
	"strings"
	"time"
)

var (
	versionFlag *string = flag.String("v", "0.9.1", "Print the version number.")
	historyUrl  *string = flag.String("url", "", "The spark history url.")
	dir         *string = flag.String("dir", filepath.Dir(os.Args[0]), "Save the markdown file.")
	tp          *string = flag.String("type", "gp", "Save type: gp(GetPage)/ps(PrintSimple)/pp(PrintPage)/ss(SaveSimple)/sp(SavePage).")
	pages               = []string{"jobs", "stages", "storage", "environment", "executors"}

	hurl string
	page string
	eurl string
)

func main() {
	flag.Parse()

	test := "http://vrv207:18080/history/application_1536025857887_0085/appattempt_1536025857887_0085_000001/jobs/"
	historyUrl = &test
	hurl = getUrl(*historyUrl)
	fmt.Printf("**Track URL:** %s\n", hurl)
	if "" == hurl {
		return
	}

	if *tp == "gp" {
		local.GetPageToLocal(test)
		//		sparder.SaveHtml(test)
		return
	}

	jobs()
	stages()
	storage()
	executors()
	environment()

	switch *tp {
	case "ps", "PrintSimple":
		sparder.PrintSimple()
	case "pp", "PrintPage":
		sparder.PrintPage()
	case "ss", "SaveSimple":
		file := createFile()
		defer file.Close()
		sparder.SaveSimple(file)
	case "sp", "SavePage":
		file := createFile()
		defer file.Close()
		sparder.SavePage(file)
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
	page = sparder.GetPage(eurl)
	jobs := sparder.BuildJobs(page)
	sparder.AddPage("Jobs", eurl, jobs)
}

func stages() {
	eurl = hurl + "/stages"
	page = sparder.GetPage(eurl)
	stages := sparder.BuildStages(page)
	sparder.AddPage("Stages", eurl, stages)
}

func storage() {
	eurl = hurl + "/storage"
	page = sparder.GetPage(eurl)
	storage := sparder.BuildStorage(page)
	sparder.AddPage("Storage", eurl, storage)
}

func executors() {
	eurl = hurl + "/executors"
	page = sparder.GetPage(eurl)
	executors := sparder.BuildExecutors(page)
	sparder.AddPage("Executors", eurl, executors)
}

func environment() {
	eurl = hurl + "/environment"
	page = sparder.GetPage(eurl)
	environment := sparder.BuildEnvironment(page)
	sparder.AddPage("Environment", eurl, environment)
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
