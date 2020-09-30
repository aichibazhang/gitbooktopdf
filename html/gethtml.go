package html

import (
	"bytes"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type GitBook struct {
	Title string
	Url   string
}

const url = `https://tsf-gitbook-1257356411.cos.ap-chengdu.myqcloud.com/1.12.4/usage/%E4%BA%A7%E5%93%81%E7%AE%80%E4%BB%8B/%E4%BA%A7%E5%93%81%E6%A6%82%E8%BF%B0.html`
const baseUrl = `https://tsf-gitbook-1257356411.cos.ap-chengdu.myqcloud.com/1.12.4/usage/`

func GetUrl() (books []*GitBook) {
	if doc, err := htmlquery.LoadURL(url); err == nil {
		list, _ := htmlquery.QueryAll(doc, "//nav/ul/li")
		for _, v := range list {
			title := strings.TrimSpace(htmlquery.InnerText(v))
			a := htmlquery.FindOne(v, "/a")
			href := htmlquery.SelectAttr(a, "href")
			var contentUrl string
			if len(href) != 0 {
				contentUrl = baseUrl + strings.Trim(href, "../")
			}
			book := GitBook{
				Title: title,
				Url:   contentUrl,
			}
			books = append(books, &book)
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

// 爬取页面
func SpiderPage(book GitBook, page chan GitBook) {
	filePath := "crawl/" + book.Title + ".html"
	client := resty.New()
	if book.Url != "" {
		_, err := client.R().SetOutput(filePath).Get(book.Url)
		if err != nil {
			return
		} else {
			fmt.Println(err)
		}
	}
	page <- book
}

// 获取文件中通用的链接
func GetHref() (string, error) {
	filePath := "crawl/2.1 Demo 工程概述.html"
	doc, err := htmlquery.LoadDoc(filePath)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	htmlHref := htmlquery.FindOne(doc, "//div[@class='book-summary']")
	htmlContent := htmlquery.OutputHTML(htmlHref, true)
	return htmlContent, nil
}
func ReplaceHtml() {
	if replaceContent, err := GetHref(); err == nil {
		fileName := make(chan string)
		files, err := ioutil.ReadDir("crawl")
		if err != nil {
			log.Fatal("html文件不存在!")
			return
		}
		for _, f := range files {
			go ReplaceFile(fileName, replaceContent, f)
		}
		for range files {
			fmt.Println(<-fileName)
		}
	}
}
func ReplaceFile(filePage chan string, replaceContent string, info os.FileInfo) {
	filePath := "crawl/" + info.Name()
	contents, _ := ioutil.ReadFile(filePath)
	newHtml := bytes.ReplaceAll(contents, []byte(replaceContent), []byte{})
	file, err := os.OpenFile(filePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}

	fmt.Println(bytes.Contains(contents, []byte(replaceContent)))
	//及时关闭file句柄
	io.WriteString(file, string(newHtml))
	filePage <- info.Name()
}
