package domain

import (
	"fmt"
	"github.com/spf13/cobra"
	"golin/global"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	spinnerChars = []string{"|", "/", "-", "\\"} //进度条更新动画
	mu           sync.Mutex
	counter      = 0 //当前已扫描的数量，计算百分比
	ch           = make(chan struct{}, 30)
	wg           = sync.WaitGroup{}
	file         = "" //读取的字典文件
	countall     = 0
	greenColor   = "\033[32m" // 设置文本颜色为绿色
	resetColor   = "\033[0m"  // 重置文本颜色
	redColor     = "\033[31m" // 设置文本颜色为红色
	size         = 100
)

func ParseFlags(cmd *cobra.Command, args []string) {
	url, _ := cmd.Flags().GetString("url")
	if url == "" {
		fmt.Printf(" [-] 域名为空！空过-u 指定！拜拜:)\n")
		return
	}
	chcount, _ := cmd.Flags().GetInt("chan") //并发数
	size, _ = cmd.Flags().GetInt("size")     //api调用的数据量
	ch = make(chan struct{}, chcount)

	file, _ = cmd.Flags().GetString("file") //读取字典文件
	if global.PathExists(file) {
		data, _ := os.ReadFile(file)
		str := strings.ReplaceAll(string(data), "\r\n", "\n")
		appurl := strings.Split(str, "\n")
		domainList = append(domainList, appurl...)
	}
	countall = len(removeDuplicates(domainList))
	if countall == 0 {
		fmt.Printf(" [-] 可碰撞子域名为0！拜拜:)\n")
		return
	}

	api, _ := cmd.Flags().GetBool("api") //读取字典文件
	if api {
		apiStart(url)
	}
	_, _ = fmt.Fprintf(colorOutput, "\n%s[*] 开始运行DNS碰撞模式 共计尝试次数:%d  并发数:%d %s\n", greenColor, countall, chcount, resetColor)

	for _, check := range removeDuplicates(domainList) {
		if len(check) == 0 {
			countall -= 1
			continue
		}
		ch <- struct{}{}
		wg.Add(1)
		if check[len(check)-1:] != "." { // 判断最后一个字符是不是 . ，不是的话则增加
			check = check + "."
		}
		go searchDomain(fmt.Sprintf("%s%s", check, url))

	}
	wg.Wait()
	time.Sleep(time.Second * 1) //等待1秒是因为并发问题，等待进度条。
	//percent()
	fmt.Print("\033[2K") // 擦除整行
	fmt.Printf("\r")

}

// removeDuplicates 切片去重
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
