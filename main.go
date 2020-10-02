package main

import (
	"gitbooktopdf/convert"
	"gitbooktopdf/crawl"
	"os"
)

func main() {
	// 获取所有html
	books:=crawl.GetUrl()
	crawl.CreateConfigJson(books)
	args := os.Args[1:]
	convert.Convert(args[0])
}
