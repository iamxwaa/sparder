package spider

import (
	"fmt"
	"os"
)

var pages []Page

type Page struct {
	title string
	url   string
	page  []TableInfo
}

var (
	CSS = `<style type="text/css">
body {
  color: #333;
}
td a,td a:hover {
  text-decoration: none;
  color: #333;
}
td div {
  text-align: center;
  border-radius: 4px;
  background-image: -moz-linear-gradient(top, #44CBFF, #34B0EE);
  background-image: -webkit-gradient(linear, 0 0, 0 100%, from(#44CBFF), to(#34B0EE));
  background-image: -webkit-linear-gradient(top, #44CBFF, #34B0EE);
  background-image: -o-linear-gradient(top, #44CBFF, #34B0EE);
  background-image: linear-gradient(to bottom, #64CBFF, #54B0EE);
  background-repeat: repeat-x;
}
.active:hover {
  color: #990000;
  cursor: pointer;
}
</style>	
	`
	JS = `<script type="text/javascript">
function show(){
    tmp = this.nextSibling
    if(tmp != undefined){
        tt = this.data
        if("" == tmp.nextSibling.style.display){
          tmp.nextSibling.style.display = "none"
          this.firstChild.innerHTML = tt +" [+]"
        }else{
          tmp.nextSibling.style.display = ""
          this.firstChild.innerHTML = tt +" [-]"
        }
    }
}

function load(){
    tb = document.getElementsByTagName("table")
    for (i in tb) {
       tb[i].style = "display:none;"
       tmp = tb[i].previousSibling
       if(tmp != undefined){
           tmp.previousSibling.data = tmp.previousSibling.firstChild.innerHTML
           tt = tmp.previousSibling.firstChild.innerHTML
           tmp.previousSibling.firstChild.innerHTML = tt +" [+]"
           tmp.previousSibling.onclick = show
           tmp.previousSibling.classList.add("active");
       }
    }
}
load()
</script>`
)

func AddPage(title string, url string, page []TableInfo) {
	p := Page{url: url, title: title, page: page}
	pages = append(pages, p)
}

func PrintSimple() {
	for _, p := range pages {
		fmt.Printf("[%s](%s)\n---\n", p.title, p.url)
		for _, p2 := range p.page {
			fmt.Printf("**%s**\n\n%s\n\n", p2.title, p2.table)
		}
	}
}

func PrintPage() {
	fmt.Println(CSS)
	for _, p := range pages {
		fmt.Printf("[%s](%s)\n---\n", p.title, p.url)
		for _, p2 := range p.page {
			fmt.Printf("**%s**\n\n%s\n\n", p2.title, p2.table)
		}
	}
	fmt.Println(JS)
}

func SaveSimple(file *os.File) {
	for _, p := range pages {
		fmt.Fprintf(file, "[%s](%s)\n---\n", p.title, p.url)
		for _, p2 := range p.page {
			fmt.Fprintf(file, "**%s**\n\n%s\n\n", p2.title, p2.table)
		}
	}
}

func SavePage(file *os.File) {
	fmt.Fprintln(file, CSS)
	for _, p := range pages {
		fmt.Fprintf(file, "[%s](%s)\n---\n", p.title, p.url)
		for _, p2 := range p.page {
			fmt.Fprintf(file, "**%s**\n\n%s\n\n", p2.title, p2.table)
		}
	}
	fmt.Fprintln(file, JS)
}
