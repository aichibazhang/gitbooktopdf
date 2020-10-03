package convert

import (
	"encoding/json"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

type Config struct {
	Charset      string   `json:"charset"`       //字符编码，默认utf-8编码
	Cover        string   `json:"cover"`         //封面图片，或者封面html文件
	Timestamp    string   `json:"date"`          //时间日期,如“2018-01-01 12:12:21”，其实是time.Time格式，但是直接用string就好
	Description  string   `json:"description"`   //摘要
	Footer       string   `json:"footer"`        //pdf的footer
	Header       string   `json:"header"`        //pdf的header
	Identifier   string   `json:"identifier"`    //即uuid，留空即可
	Language     string   `json:"language"`      //语言，如zh、en、zh-CN、en-US等
	Creator      string   `json:"creator"`       //作者，即author
	Publisher    string   `json:"publisher"`     //出版单位
	Contributor  string   `json:"contributor"`   //同Publisher
	Title        string   `json:"title"`         //文档标题
	Format       string   `json:"format"`        //导出格式，可选值：pdf、epub、mobi
	FontSize     string   `json:"font_size"`     //默认的pdf导出字体大小
	PaperSize    string   `json:"paper_size"`    //页面大小
	MarginLeft   string   `json:"margin_left"`   //PDF文档左边距，写数字即可，默认72pt
	MarginRight  string   `json:"margin_right"`  //PDF文档左边距，写数字即可，默认72pt
	MarginTop    string   `json:"margin_top"`    //PDF文档左边距，写数字即可，默认72pt
	MarginBottom string   `json:"margin_bottom"` //PDF文档左边距，写数字即可，默认72pt
	More         []string `json:"more"`          //更多导出选项[PDF导出选项，具体参考：https://manual.calibre-ebook.com/generated/en/ebook-convert.html#pdf-output-options]
	Toc          []Toc    `json:"toc"`           //目录
	Link         []string `json:"-"`             //这个不需要赋值
}
type WebSite struct {
	Url     string `yaml:"url"`      // 爬取的gitbook url
	BaseUrl string `yaml:"base_url"` // 爬取的gitbook baseUrl
}
type Article struct {
	Title string `yaml:"title"`
}
type ConfigFile struct {
	WebSite WebSite `yaml:"website"`
	Article Article `yaml:"article"`
}

//目录结构
type Toc struct {
	Id    int    `json:"id"`
	Link  string `json:"link"`
	Pid   int    `json:"pid"`
	Title string `json:"title"`
}

//media-type
var MediaType = map[string]string{
	".jpeg":  "image/jpeg",
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".gif":   "image/gif",
	".ico":   "image/x-icon",
	".bmp":   "image/bmp",
	".html":  "application/xhtml+xml",
	".xhtml": "application/xhtml+xml",
	".htm":   "application/xhtml+xml",
	".otf":   "application/x-font-opentype",
	".ttf":   "application/x-font-ttf",
	".js":    "application/x-javascript",
	".ncx":   "x-dtbncx+xml",
	".txt":   "text/plain",
	".xml":   "text/xml",
	".css":   "text/css",
}

//根据文件扩展名，获取media-type
func GetMediaType(ext string) string {
	if mt, ok := MediaType[strings.ToLower(ext)]; ok {
		return mt
	}
	return ""
}

//解析配置文件
func parseConfig(configFile string) (cfg Config, err error) {
	var b []byte
	if b, err = ioutil.ReadFile(configFile); err == nil {
		err = json.Unmarshal(b, &cfg)
	}
	return
}

// 获取初始配置
func InitConfig() (cfg ConfigFile, err error) {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
		return cfg, err
	}
	yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return cfg, err
	}
	return
}
