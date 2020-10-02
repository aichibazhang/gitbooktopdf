package crawl

import (
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

type GitBook struct {
	Title string `json:"title"`
	Id    int    `json:"id"`
	Pid   int    `json:"pid"`
	Link  string `json:"link"`
}

var tocs map[int]string

const url = `https://tsf-gitbook-1257356411.cos.ap-chengdu.myqcloud.com/1.12.4/usage/%E4%BA%A7%E5%93%81%E7%AE%80%E4%BB%8B/%E4%BA%A7%E5%93%81%E6%A6%82%E8%BF%B0.html`
const baseUrl = `https://tsf-gitbook-1257356411.cos.ap-chengdu.myqcloud.com/1.12.4/usage/`
const folder = `crawl/input/`

func GetUrl() (books []*GitBook) {
	if doc, err := htmlquery.LoadURL(url); err == nil {
		var pids []int
		tocs = make(map[int]string)
		// 获取所有header,标题
		headers, _ := htmlquery.QueryAll(doc, "//li[@class='header']")
		for index := range headers {
			header := htmlquery.InnerText(headers[index])
			headerIndex, _ := strconv.Atoi(header)
			tocContent := strings.Split(header, ".")
			tocIndex, _ := strconv.Atoi(tocContent[0])
			tocs[tocIndex] = header
			pids = append(pids, headerIndex)
		}
		list, _ := htmlquery.QueryAll(doc, "//nav/ul/li")
		for _, v := range list {
			var (
				book      GitBook
				content   string
				levelData string
			)
			title := strings.TrimSpace(htmlquery.InnerText(v))
			level := htmlquery.FindOne(v, "/@data-level")
			if level != nil {
				levelData = htmlquery.InnerText(level)
				content = strings.ReplaceAll(levelData, ".", "")
				id, _ := strconv.Atoi(content)
				book.Title, book.Id = title, id
			}
			a := htmlquery.FindOne(v, "/a")
			href := htmlquery.SelectAttr(a, "href")
			var contentUrl string
			if len(href) != 0 {
				if strings.Contains(href, "../") {
					contentUrl = baseUrl + strings.Trim(href, "../")
				} else if strings.Contains(href, "./") {
					contentUrl = baseUrl + strings.Trim(href, "./")
				}
			}
			book.Link = contentUrl
			if contentIndex, err := strconv.Atoi(strings.Split(levelData, ".")[0]); err == nil {
				book.Pid = contentIndex - 1
			}
			if len(book.Link) != 0 {
				books = append(books, &book)
			}
		}
	}
	return
}
func CrawlUrl(urls []*GitBook) {
	page := make(chan GitBook)
	for _, uri := range urls {
		go SpiderPage(*uri, page)
	}

	for i := 0; i < len(urls); i++ {
		fmt.Println(<-page)
	}
}
func CrawlSummary() {
	toc := make(chan string)
	for index, title := range tocs {
		go SpiderSummary(index, title, toc)
	}
	for i := 0; i < len(tocs); i++ {
		fmt.Println(<-toc)
	}
}

// 爬取页面
func SpiderPage(book GitBook, page chan GitBook) {
	if doc, err := htmlquery.LoadURL(book.Link); err == nil {
		body := htmlquery.Find(doc, "//div[@class='page-inner']")
		if len(body) != 0 {
			pdfBody := body[0]
			htmlBody := htmlquery.OutputHTML(pdfBody, true)
			htmlTempleta := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
    <title>%v</title>
    <link href="gitbook.css" rel="stylesheet">
</head>
<body>%v
</body>
</html>`
			htmlTempleta = fmt.Sprintf(htmlTempleta, book.Title, htmlBody)
			ioutil.WriteFile(folder+strconv.Itoa(book.Id)+".html", []byte(htmlTempleta), os.ModePerm)
		}
	}
	page <- book
}
func SpiderSummary(key int, value string, page chan string) {
	htmlTempleta := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
    <title>%v</title>
    <link href="gitbook.css" rel="stylesheet">
</head>
<body>%v
</body>
</html>`
	htmlTempleta = fmt.Sprintf(htmlTempleta, value, value)
	ioutil.WriteFile(folder+strconv.Itoa(key)+".html", []byte(htmlTempleta), os.ModePerm)
	page <- value
}
func CreateConfigJson(books []*GitBook) {
	for i := range books {
		book := books[i]
		if len(book.Link) != 0 {
			books[i].Link = strconv.Itoa(books[i].Id) + ".html"
		}
	}
	var keys []int
	for k := range tocs {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, index := range keys {
		book := GitBook{
			Title: tocs[index],
			Id:    index,
			Pid:   0,
			Link:  strconv.Itoa(index) + ".html",
		}
		books = append(books, &book)
	}
	b, err := json.Marshal(books)
	if err != nil {
		fmt.Println("Umarshal failed:", err)
		return
	}
	jsonTemplate :=
		`{
  "charset": "utf-8",
  "cover": "",
  "date": "",
  "description": "diaosi.love 程序员福利:免费翻墙,实用工具,你值得拥有!!!",
  "footer": "<p style='color:#8E8E8E;font-size:12px;'>本文档由 <a href='https://www.diaosi.love' style='text-decoration:none;color:#1abc9c;font-weight:bold;'>福利工具(diaosi.love)</a> 构建 <span style='float:right'>- _PAGENUM_ -</span></p>",
  "header": "<p style='color:#8E8E8E;font-size:12px;'>_SECTION_</p>",
  "identifier": "",
  "language": "zh-CN",
  "creator": "福利(diaosi.love)",
  "publisher": "福利(diaosi.love)",
  "contributor": "福利(diaosi.love)",
  "title": "腾讯微服务平台开发文档",
  "format": "pdf",
  "font_size": "13",
  "paper_size": "a4",
  "margin_left": "72",
  "margin_right": "72",
  "margin_top": "72",
  "margin_bottom": "72",
  "more": [],
  "toc": %v
}`
	jsonTemplate = fmt.Sprintf(jsonTemplate, string(b))
	ioutil.WriteFile(folder+"test.json", []byte(jsonTemplate), os.ModePerm)
}
