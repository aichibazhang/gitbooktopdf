# 缘由
一些不规范的gitbook,想要将其转成pdf文件找了很多工具都无法完成
> 参考此链接: https://tsf-gitbook-1257356411.cos.ap-chengdu.myqcloud.com/1.12.4/usage/%E4%BA%A7%E5%93%81%E7%AE%80%E4%BB%8B/%E4%BA%A7%E5%93%81%E6%A6%82%E8%BF%B0.html
# 不规范链接说明
1. 点击不同的章节,可以获取的url不同
2. 部分url用js隐藏,目前暂未找到很好的解决方法
3. 部分链接无法访问,比如文档说明那部分链接
4. 根路径无法访问(一般gitbook爬取工具都是从根路径爬取)
# 前提
- 安装calibre
- 下载地址：https://calibre-ebook.com/download
- 根据自己的系统安装对应的calibre（需要注意的是，calibre要安装3.x版本的，2.x版本的功能不是很强大。反正安装最新的就好。）
安装完calibre之后，将calibre加入到系统环境变量中，执行下面的命令之后显示3.x的版本即表示安装成功。

```ebook-convert --version```
# 如何使用
1. 生成json文件,运行main中`generatorHtmlAndJson()`
2. 修改生成的json文件,删除部分json文件,根据实际场景设置,比如腾讯开发文档那个链接因为文档说明
那个链接无法点击,所以我们需要将这部分相关json删除
3. 运行main中`toPdf()`
