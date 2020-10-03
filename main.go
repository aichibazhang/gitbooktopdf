package main

import (
	"gitbooktopdf/convert"
	"gitbooktopdf/crawl"
	"sync"
)

func generatorHtmlAndJson() {
	var wg sync.WaitGroup
	cfg, _ := convert.InitConfig()
	books := crawl.GetUrl(cfg)
	crawl.CrawlUrl(books, wg)
	crawl.CrawlSummary(wg)
	wg.Wait()
	crawl.CreateConfigJson(books, cfg)
}
func toPdf() {
	configPath := "crawl/input/config.json"
	convert.Convert(configPath)
}
func main() {
	// 分两步走:1. 生成json文件及html源文件 2. 调整json文件并生成pdf
	//generatorHtmlAndJson()
	toPdf()

}
