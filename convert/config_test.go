package convert

import (
	"fmt"
	"strings"
	"testing"
)

func TestConfigParse(t *testing.T) {
	cfg, err := parseConfig("../example/gogs_zh/config.json")
	if err != nil {
		t.Log(err, "config文件解析错误,请检查config")
		return
	}
	fmt.Println(cfg)
	str:=[]string{"tst","test"}
	fmt.Println(strings.Join(str,""))
}
