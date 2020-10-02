package convert

import (
	"fmt"
	util "gitbooktopdf/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 参考链接: https://manual.calibre-ebook.com/generated/en/ebook-convert.html
var (
	output       = "output"        //文档导出文件夹
	ebookConvert = "ebook-convert" //导出命令行
)

type Converter struct {
	BasePath       string
	Config         Config
	GeneratedCover string
}

func NewConverter(configFile string) (convert *Converter, err error) {
	fmt.Println("配置路径为:",configFile)
	cfg, err := parseConfig(configFile)
	if err != nil {
		fmt.Println(err, "config文件解析错误,请检查config")
		return
	}
	filePath, err := filepath.Abs(filepath.Dir(configFile))
	if err != nil {
		fmt.Println(err, "文件路径错误")
		return
	}
	if len(cfg.Timestamp) == 0 {
		cfg.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}
	if len(cfg.Charset) == 0 {
		cfg.Charset = "utf-8"
	}
	convert = &Converter{
		Config:   cfg,
		BasePath: filePath,
	}
	return
}
func Convert(configPath string) {
	convert, _ := NewConverter(configPath)
	defer convert.converterDefer()
	if err := convert.mimeType(); err != nil {
		return
	}
	if err := convert.container(); err != nil {
		return
	}
	if err := convert.tocNcx(); err != nil {
		return
	}
	// 生成首页
	if err := convert.home(); err != nil {
		return
	}
	if err := convert.generateTitlePage(); err != nil {
		return
	}
	if err := convert.generateContentOpf(); err != nil {
		return
	}
	//将当前文件夹下的所有文件压缩成zip包，然后直接改名成content.epub
	f := convert.BasePath + "/content.epub"
	os.Remove(f)
	if err := util.Zip(f, convert.BasePath); err == nil {
		//创建导出文件夹
		os.Mkdir(convert.BasePath+"/"+output, os.ModePerm)
		//处理直接压缩后的content.epub文件，将其转成新的content-tmp.epub文件
		//再用content-tmp.epub文件替换掉content.epub文件，这样content.epub文件转PDF的时候，就不会出现空白页的情况了
		tmp := convert.BasePath + "/content-tmp.epub"
		if err = exec.Command(ebookConvert, f, tmp).Run(); err != nil {
			return
		} else {
			os.Remove(f)
			os.Rename(tmp, f)
		}
		if err := convert.convertToPdf(); err != nil {
			fmt.Println(err)
		}
		return
	}
}
func (this *Converter) generateContentOpf() (err error) {
	var (
		guide       string
		manifest    string
		manifestArr []string
		spine       string //注意：如果存在封面，则需要把封面放在第一个位置
		spineArr    []string
	)

	meta := `<dc:title>%v</dc:title>
			<dc:contributor opf:role="bkp">%v</dc:contributor>
			<dc:publisher>%v</dc:publisher>
			<dc:description>%v</dc:description>
			<dc:language>%v</dc:language>
			<dc:creator opf:file-as="Unknown" opf:role="aut">%v</dc:creator>
			<meta name="calibre:timestamp" content="%v"/>
	`
	meta = fmt.Sprintf(meta, this.Config.Title, this.Config.Contributor, this.Config.Publisher, this.Config.Description, this.Config.Language, this.Config.Creator, this.Config.Timestamp)
	if len(this.Config.Cover) > 0 {
		meta = meta + `<meta name="cover" content="cover"/>`
		guide = `<reference href="titlepage.xhtml" title="Cover" type="cover"/>`
		manifest = fmt.Sprintf(`<item href="%v" id="cover" media-type="%v"/>`, this.Config.Cover, GetMediaType(filepath.Ext(this.Config.Cover)))
		spineArr = append(spineArr, `<itemref idref="titlepage"/>`)
	}

	if _, err := os.Stat(this.BasePath + "/summary.html"); err == nil {
		spineArr = append(spineArr, `<itemref idref="summary"/>`) //目录

	}

	//扫描所有文件
	if files, err := util.ScanFiles(this.BasePath); err == nil {
		basePath := strings.Replace(this.BasePath, "\\", "/", -1)
		for _, file := range files {
			if !file.IsDir {
				ext := strings.ToLower(filepath.Ext(file.Path))
				sourcefile := strings.TrimPrefix(file.Path, basePath+"/")
				id := "ncx"
				if ext != ".ncx" {
					if file.Name == "titlepage.xhtml" { //封面
						id = "titlepage"
					} else if file.Name == "summary.html" { //目录
						id = "summary"
					} else {
						id = util.Md5Crypt(sourcefile)
					}
				}
				if mt := GetMediaType(ext); mt != "" { //不是封面图片，且media-type不为空
					if sourcefile != strings.TrimLeft(this.Config.Cover, "./") { //不是封面图片，则追加进来。封面图片前面已经追加进来了
						manifestArr = append(manifestArr, fmt.Sprintf(`<item href="%v" id="%v" media-type="%v"/>`, sourcefile, id, mt))
					}
				}
			}
		}

		items := make(map[string]string)
		for _, link := range this.Config.Link {
			id := util.Md5Crypt(link)
			if _, ok := items[id]; !ok { //去重
				items[id] = id
				spineArr = append(spineArr, fmt.Sprintf(`<itemref idref="%v"/>`, id))
			}
		}
		manifest = manifest + strings.Join(manifestArr, "\n")
		spine = strings.Join(spineArr, "\n")
	} else {
		return err
	}

	pkg := `<?xml version='1.0' encoding='` + this.Config.Charset + `'?>
		<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uuid_id" version="2.0">
		  <metadata xmlns:opf="http://www.idpf.org/2007/opf" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:calibre="http://calibre.kovidgoyal.net/2009/metadata">
			%v
		  </metadata>
		  <manifest>
			%v
		  </manifest>
		  <spine toc="ncx">
			%v
		  </spine>
			%v
		</package>
	`
	if len(guide) > 0 {
		guide = `<guide>` + guide + `</guide>`
	}
	pkg = fmt.Sprintf(pkg, meta, manifest, spine, guide)
	return ioutil.WriteFile(this.BasePath+"/content.opf", []byte(pkg), os.ModePerm)
}

//生成封面
func (this *Converter) generateTitlePage() (err error) {
	if ext := strings.ToLower(filepath.Ext(this.Config.Cover)); !(ext == ".html" || ext == ".xhtml") {
		xml := `<?xml version='1.0' encoding='` + this.Config.Charset + `'?>
				<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="` + this.Config.Language + `">
					<head>
						<meta http-equiv="Content-Type" content="text/html; charset=` + this.Config.Charset + `"/>
						<meta name="calibre:cover" content="true"/>
						<title>Cover</title>
						<style type="text/css" title="override_css">
							@page {padding: 0pt; margin:0pt}
							body { text-align: center; padding:0pt; margin: 0pt; }
						</style>
					</head>
					<body>
						<div>
							<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="100%" height="100%" viewBox="0 0 800 1068" preserveAspectRatio="none">
								<image width="800" height="1068" xlink:href="` + strings.TrimPrefix(this.Config.Cover, "./") + `"/>
							</svg>
						</div>
					</body>
				</html>
		`
		if err = ioutil.WriteFile(this.BasePath+"/titlepage.xhtml", []byte(xml), os.ModePerm); err == nil {
			this.GeneratedCover = "titlepage.xhtml"
		}
	}
	return
}
func (this *Converter) mimeType() error {
	return ioutil.WriteFile(this.BasePath+"/mimetype", []byte("application/epub+zip"), os.ModePerm)
}
func (this *Converter) container() (err error) {
	xml := `<?xml version="1.0"?>
			<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
			   <rootfiles>
				  <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
			   </rootfiles>
			</container>
    `
	folder := this.BasePath + "/META-INF"
	os.MkdirAll(folder, os.ModePerm)
	err = ioutil.WriteFile(folder+"/container.xml", []byte(xml), os.ModePerm)
	return
}
func (this *Converter) convertToPdf() error {
	args := []string{
		this.BasePath+"/content.epub",
		this.BasePath + "/" + output + "/book.pdf",
	}
	//页面大小
	if len(this.Config.PaperSize) > 0 {
		args = append(args, "--paper-size", this.Config.PaperSize)
	}
	//文字大小
	if len(this.Config.FontSize) > 0 {
		args = append(args, "--pdf-default-font-size", this.Config.FontSize)
	}

	//header template
	if len(this.Config.Header) > 0 {
		args = append(args, "--pdf-header-template", this.Config.Header)
	}

	//footer template
	if len(this.Config.Footer) > 0 {
		args = append(args, "--pdf-footer-template", this.Config.Footer)
	}

	if len(this.Config.MarginLeft) > 0 {
		args = append(args, "--pdf-page-margin-left", this.Config.MarginLeft)
	}
	if len(this.Config.MarginTop) > 0 {
		args = append(args, "--pdf-page-margin-top", this.Config.MarginTop)
	}
	if len(this.Config.MarginRight) > 0 {
		args = append(args, "--pdf-page-margin-right", this.Config.MarginRight)
	}
	if len(this.Config.MarginBottom) > 0 {
		args = append(args, "--pdf-page-margin-bottom", this.Config.MarginBottom)
	}

	//更多选项
	if len(this.Config.More) > 0 {
		args = append(args, this.Config.More...)
	}
	fmt.Println(args)
	cmd := exec.Command(ebookConvert, args...)
	return cmd.Run()
}

func (this *Converter) tocNcx() error {
	ncx := `<?xml version='1.0' encoding='` + this.Config.Charset + `'?>
			<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1" xml:lang="%v">
			  <head>
				<meta content="4" name="dtb:depth"/>
				<meta content="calibre (2.85.1)" name="dtb:generator"/>
				<meta content="0" name="dtb:totalPageCount"/>
				<meta content="0" name="dtb:maxPageNumber"/>
			  </head>
			  <docTitle>
				<text>%v</text>
			  </docTitle>
			  <navMap>%v</navMap>
			</ncx>
	`
	codes, _ := this.tocToXml(0, 1)
	ncx = fmt.Sprintf(ncx, this.Config.Language, this.Config.Title, strings.Join(codes, ""))
	return ioutil.WriteFile(this.BasePath+"/toc.ncx", []byte(ncx), os.ModePerm)
}
func (this *Converter) tocToXml(pid, idx int) (codes []string, next_idx int) {
	var code string
	for _, toc := range this.Config.Toc {
		if toc.Pid == pid {
			code, idx = this.getNavPoint(toc, idx)
			codes = append(codes, code)
			for _, item := range this.Config.Toc {
				if item.Pid == toc.Id {
					code, idx = this.getNavPoint(item, idx)
					codes = append(codes, code)
					var code_arr []string
					code_arr, idx = this.tocToXml(item.Id, idx)
					codes = append(codes, code_arr...)
					codes = append(codes, `</navPoint>`)
				}
			}
			codes = append(codes, `</navPoint>`)
		}
	}
	next_idx = idx
	return
}

//删除生成导出文档而创建的文件
func (this *Converter) converterDefer() {
	//删除不必要的文件
	os.RemoveAll(this.BasePath + "/META-INF")
	os.RemoveAll(this.BasePath + "/content.epub")
	os.RemoveAll(this.BasePath + "/mimetype")
	os.RemoveAll(this.BasePath + "/toc.ncx")
	os.RemoveAll(this.BasePath + "/content.opf")
	os.RemoveAll(this.BasePath + "/titlepage.xhtml") //封面图片待优化
	os.RemoveAll(this.BasePath + "/summary.html")    //文档目录
}

//生成navPoint
func (this *Converter) getNavPoint(toc Toc, idx int) (navpoint string, nextidx int) {
	navpoint = `
	<navPoint id="id%v" playOrder="%v">
		<navLabel>
			<text>%v</text>
		</navLabel>
		<content src="%v"/>`
	navpoint = fmt.Sprintf(navpoint, toc.Id, idx, toc.Title, toc.Link)
	this.Config.Link = append(this.Config.Link, toc.Link)
	nextidx = idx + 1
	return
}

func (this *Converter) home() error {
	//目录
	summary := `<!DOCTYPE html>
				<html lang="` + this.Config.Language + `">
				<head>
				    <meta charset="` + this.Config.Charset + `">
				    <title>目录</title>
				    <style>
				        body{margin: 0px;padding: 0px;}h1{text-align: center;padding: 0px;margin: 0px;}ul,li{list-style: none;}ul{padding-left:0px;}li>ul{padding-left: 2em;}
				        a{text-decoration: none;color: #4183c4;text-decoration: none;font-size: 16px;line-height: 28px;}
				    </style>
				</head>
				<body>
				    <h1>目&nbsp;&nbsp;&nbsp;&nbsp;录</h1>
				    %v
				</body>
				</html>`
	summary = fmt.Sprintf(summary, strings.Join(this.tocToSummary(0), ""))
	return ioutil.WriteFile(this.BasePath+"/summary.html", []byte(summary), os.ModePerm)
}

//将toc转成summary目录
func (this *Converter) tocToSummary(pid int) (summarys []string) {
	summarys = append(summarys, "<ul>")
	for _, toc := range this.Config.Toc {
		if toc.Pid == pid {
			summarys = append(summarys, fmt.Sprintf(`<li><a href="%v">%v</a></li>`, toc.Link, toc.Title))
			for _, item := range this.Config.Toc {

				if item.Pid == toc.Id {
					summarys = append(summarys, fmt.Sprintf(`<li><ul><li><a href="%v">%v</a></li>`, item.Link, item.Title))
					summarys = append(summarys, "<li>")
					summarys = append(summarys, this.tocToSummary(item.Id)...)
					summarys = append(summarys, "</li></ul></li>")
				}

			}
		}
	}
	summarys = append(summarys, "</ul>")
	return
}
