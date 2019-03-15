package spider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type ExpInfo struct {
	p     string
	old   string
	ns    string
	index int
	wrap  string
}

type TableInfo struct {
	table string
	title string
}

const (
	RP      = "$VALUE$"
	RP1     = "$VALUE1$"
	WRAP_P  = "<div>" + RP + "</div>"
	WRAP_A  = "<a title=\"" + RP1 + "\">" + RP + "</a>"
	MAX_LEN = 67
)

var (
	TDEXP        = ExpInfo{p: ">([^\\<]+)</td", index: 1}
	TDEXP_WRAP   = ExpInfo{p: ">([^\\<]+)</td", wrap: WRAP_P, index: 1}
	TIMEEXP      = ExpInfo{p: ">([^\\<]+)</td", old: ":", ns: "&#58;", index: 1}
	AEXP         = ExpInfo{p: ">([^\\<]+)</a", index: 1}
	SPANEXP      = ExpInfo{p: ">([^\\<]+)</span", index: 1}
	SPANEXP_WRAP = ExpInfo{p: ">([^\\<]+)</span", wrap: WRAP_P, index: 1}

	TBody            = ExpInfo{p: "<tbody>([\\n\\s\\S]*)</tbody>", index: 1}
	EnvironmentTBody = ExpInfo{p: "<tbody>(\\n|\\r\\n)[\\s]*.*(\\n|\\r\\n).*((\\n|\\r\\n)[\\s]*)?</tbody>", index: 0}
	JobsTBody        = ExpInfo{p: "<tbody>([\\s\\n]+.*[\\s\\n]+)+</tbody>", index: 0}
	StagesTBody      = ExpInfo{p: "<tbody>([\\n\\s\\S]*)</tbody>", index: 1}
)

func Test() {
	fmt.Println("")
}

func GetPage(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b)
}

func BuildExecutors(p string) []TableInfo {
	heads := make([][]string, 1)
	tmpHead1 := getMatchValues(p, "<th[^\\>]*?>([^\\<]*)</th>", 1)
	tmpHead2 := getMatchValues(p, "<span[^\\>]*?>([^\\<]*)</span>", 1)
	heads[0] = []string{
		tmpHead1[0], tmpHead1[1], tmpHead1[2], tmpHead2[1], tmpHead1[3],
		tmpHead1[4], tmpHead1[5], tmpHead1[6], tmpHead1[7], tmpHead1[8],
		tmpHead2[2], tmpHead2[3], tmpHead2[4]} //, tmpHead1[9] }
	cexps := make([][]ExpInfo, 1)
	cexps[0] = []ExpInfo{TDEXP, TDEXP, TDEXP, TDEXP, TDEXP,
		TDEXP, TDEXP, TDEXP, TDEXP, TDEXP,
		TDEXP, TDEXP, TDEXP} //, TDEXP}
	return buildTableInfo(Build(p, TBody, heads, cexps), "Executor Table")
}

func BuildEnvironment(p string) []TableInfo {
	heads := make([][]string, 4)
	tmpHead := getMatchValues(p, "<th[^\\<]*?>([^\\<]*)</th>", 1)
	heads[0] = []string{tmpHead[0], tmpHead[1]}
	heads[1] = []string{tmpHead[2], tmpHead[3]}
	heads[2] = []string{tmpHead[4], tmpHead[5]}
	heads[3] = []string{tmpHead[6], tmpHead[7]}
	cexps := make([][]ExpInfo, 4)
	cexps[0] = []ExpInfo{TDEXP, TDEXP}
	cexps[1] = []ExpInfo{TDEXP, TDEXP}
	cexps[2] = []ExpInfo{TDEXP, TDEXP}
	cexps[3] = []ExpInfo{TDEXP, TDEXP}
	return buildTableInfo(Build(p, EnvironmentTBody, heads, cexps), "Runtime Information", "Spark Properties", "System Properties", "Classpath Entries")
}

func BuildStorage(p string) []TableInfo {
	heads := make([][]string, 1)
	heads[0] = getMatchValues(p, "<th[^\\<]*?>([^\\<]*)</th>", 1)
	cexps := make([][]ExpInfo, 1)
	cexps[0] = []ExpInfo{AEXP, TDEXP, TDEXP, TDEXP, TDEXP, TDEXP, TDEXP}
	return buildTableInfo(Build(p, TBody, heads, cexps), "Storage Table")
}

func BuildStages(p string) []TableInfo {
	h1 := getMatchValues(p, "<th[^\\<]*?>([^\\>]*)</th>", 1)
	h2 := getMatchValues(p, "<th><span[^\\<]*?>([^\\>]*)</span>", 1)
	h3 := getMatchValues(p, "<span data-toggle=\"tooltip\" data-placement=\"left\" title=\"[^\\>]*\">([^\\<]*\n[^\\<]*)*</span>", 1)
	head := union(h1, h2, h3)
	var heads [][]string
	var cexps [][]ExpInfo
	cexps = append(cexps, []ExpInfo{TDEXP, AEXP, TIMEEXP, TDEXP, SPANEXP_WRAP, TDEXP, TDEXP, TDEXP, TDEXP, TDEXP})
	if len(head) > 9 {
		heads = append(heads, []string{head[0], head[1], head[2], head[3], head[4], head[11], head[12], head[13], head[17]})
		heads = append(heads, []string{head[5], head[6], head[7], head[8], head[9], head[14], head[15], head[16], head[18], head[10]})
	} else {
		heads = append(heads, head)
	}

	var ss []string
	tbodys := getTbodys(p, StagesTBody)
	if len(tbodys) > 0 {
		trtds := getTrTds(tbodys[0])
		vs := getValueFromTrTds(trtds, cexps[0])
		start := 0
		if len(vs) > 1 && len(heads) > 1 {
			for i := 1; i < len(vs); i++ {
				id, _ := strconv.Atoi(vs[i][0])
				if id == 0 {
					start = i + 1
					ss = append(ss, markDownTable(heads[0], vs[0:start]))
					break
				}
			}
			ss = append(ss, markDownTable(heads[1], vs[start:]))
		} else {
			ss = append(ss, markDownTable(heads[0], vs))
		}
	}

	return buildTableInfo(ss, "Stage Table", "Stage Fail Table")
}

func BuildJobs(p string) []TableInfo {
	var heads [][]string
	var cexps [][]ExpInfo
	tmpHead := getMatchValues(p, "<th[^\\<]*?>([^\\<]*)</th>", 1)
	if len(tmpHead) > 6 {
		heads = append(heads, []string{tmpHead[0], tmpHead[1], tmpHead[2], tmpHead[3], tmpHead[4], tmpHead[5]})
		heads = append(heads, []string{tmpHead[6], tmpHead[7], tmpHead[8], tmpHead[9], tmpHead[10], tmpHead[11]})
		cexps = append(cexps, []ExpInfo{TDEXP, AEXP, TIMEEXP, TDEXP, TDEXP_WRAP, SPANEXP_WRAP})
		cexps = append(cexps, []ExpInfo{TDEXP, AEXP, TIMEEXP, TDEXP, TDEXP_WRAP, SPANEXP_WRAP})
	} else {
		heads = append(heads, tmpHead)
		cexps = append(cexps, []ExpInfo{TDEXP, AEXP, TIMEEXP, TDEXP, TDEXP_WRAP, SPANEXP_WRAP})
	}
	return buildTableInfo(Build(p, JobsTBody, heads, cexps), "Job Table", "Job Fail Table")
}

func Build(p string, bexp ExpInfo, heads [][]string, cexps [][]ExpInfo) []string {
	tbodys := getTbodys(p, bexp)
	s := make([]string, len(tbodys))
	for i, tbody := range tbodys {
		if len(heads[i]) < 2 {
			continue
		}
		trtds := getTrTds(tbody)
		vs := getValueFromTrTds(trtds, cexps[i])
		s[i] = markDownTable(heads[i], vs)
	}
	return s
}

func markDownTable(head []string, arr [][]string) string {
	//构造表头
	hlen := len(head)
	var tmp []string
	if len(arr) > 0 {
		tmp = make([]string, len(arr)+2)
	} else {
		tmp = make([]string, 2)
	}
	tmp[0] = strings.Join(head, "|")
	line := make([]string, hlen)
	for i := 0; i < hlen; i++ {
		line[i] = "--"
	}
	tmp[1] = strings.Join(line, "|")
	for i, a1 := range arr {
		tmp[i+2] = strings.Join(a1[:hlen], "|")
	}
	return strings.Join(tmp, "\n")
}

func getValueFromTrTds(trtds [][]string, cexp []ExpInfo) [][]string {
	values := make([][]string, len(trtds))
	for i, tds := range trtds {
		values[i] = make([]string, len(cexp))
		for j, e := range cexp {
			if "" == e.p {
				values[i][j] = "-"
			} else {
				if j >= len(tds) {
					break
				}
				vv := getMatchValues(tds[j], e.p, e.index)[0]
				if "" != e.old {
					vv = strings.Replace(vv, e.old, e.ns, -1)
				}
				if "" != e.wrap {
					tvv := e.wrap
					vv = strings.Replace(tvv, RP, vv, 1)
				}
				if len(vv) > MAX_LEN {
					tvv := WRAP_A
					tvv = strings.Replace(tvv, RP1, vv, 1)
					vv = vv[:MAX_LEN] + "..."
					vv = strings.Replace(tvv, RP, vv, 1)
				}
				values[i][j] = vv
			}
		}
	}
	return values
}

func getTbodys(p string, bexp ExpInfo) []string {
	tbody := getMatchValues(p, bexp.p, bexp.index)
	if "-" == tbody[0] {
		return make([]string, 0)
	}
	for i, t := range tbody {
		tbody[i] = strings.Replace(t, "</tbody>", "", 1)
	}
	return tbody
}

func getTrTds(p string) [][]string {
	if p == "-" {
		return make([][]string, 0)
	}
	np := strings.Replace(p, "\n", "", -1)
	np = strings.Replace(np, "</tr>", "</tr>\n", -1)
	nps := strings.Split(np, "\n")
	if "" == strings.TrimSpace(nps[len(nps)-1]) {
		nps = nps[:len(nps)-1]
	}
	trtds := make([][]string, len(nps))
	for i, tr := range nps {
		ttr := strings.Replace(tr, "</td>", "</td>\n", -1)
		ttrs := strings.Split(ttr, "\n")
		if "" == ttrs[len(ttrs)-1] {
			ttrs = ttrs[:len(ttrs)-1]
		}
		trtds[i] = ttrs
	}

	return trtds
}

func getMatchValues(p string, exp string, index int) []string {
	r, _ := regexp.Compile(exp)
	tmp := r.FindAllStringSubmatch(p, -1)
	arr := make([]string, len(tmp))
	for i := 0; i < len(arr); i++ {
		arr[i] = strings.TrimSpace(tmp[i][index])
	}
	if 0 == len(arr) {
		return []string{"-"}
	}
	return arr
}

func replace(s []string, old string, ns string) {
	for i := 0; i < len(s); i++ {
		s[i] = strings.Replace(s[i], old, ns, -1)
	}
}

func union(arr ...[]string) []string {
	l := 0
	for i := 0; i < len(arr); i++ {
		l += len(arr[i])
	}
	s := make([]string, l)
	m := 0
	for i := 0; i < len(arr) && m < l; i++ {
		for j := 0; j < len(arr[i]); j++ {
			s[m] = arr[i][j]
			m++
		}
	}

	return s
}

func buildTableInfo(ss []string, title ...string) []TableInfo {
	ti := make([]TableInfo, len(ss))
	l := len(title) - 1
	for i, j := range ss {
		if i > l {
			ti[i] = TableInfo{table: j, title: title[l] + " " + strconv.Itoa(i+1)}
		} else {
			ti[i] = TableInfo{table: j, title: title[i]}
		}
	}
	return ti
}
