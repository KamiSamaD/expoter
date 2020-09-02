package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"time"
)

type itemData struct {
	name     string
	itemType string
	value    float64
	lable    string
	lablename  string
}





//
func allOutput()  map[int]itemData{
	return RunScripts(ListScripts(path))
}
//获取文件列表
func ListScripts(path string) (scriptsname []string) {
	var cmd *exec.Cmd
	path1 := path + "/scripts"
	cmd = exec.Command("/bin/ls", path1)
	file, _ := cmd.Output()
	scriptsname = strings.Fields(string(file))
	fmt.Printf("获取文件列表 %T\n%v\n", scriptsname, scriptsname)
	return scriptsname

}

//执行脚本文件，并获取输出
func RunScripts(scriptsname []string)  map[int]itemData {
	var j string
	var promPATH string
	var item itemData
	var cmd2 *exec.Cmd
	//var result []byte
	var cmd3 *exec.Cmd
	var result2 []byte
	var s []string
	var g int
	itemList := make(map[int]itemData)
	for l, i := range scriptsname {
		//执行脚本，将会在./prom中生成一个prom文件
		j = path + "/scripts/" + i
		cmd2 = exec.Command("/bin/bash", j)
		//result, _ = cmd2.Output()
		//fmt.Printf("scripts output:  %v", string(result))
		_, err := cmd2.Output()
		if err != nil {
			fmt.Printf("\n执行脚本报错\n 错误=%v", err)
		}
		//获取生成的prom文件中的数据
		promPATH = path + "/prom/" + i + ".prom"
		//fmt.Println(promPATH)
		cmd3 = exec.Command("/bin/cat", promPATH)
		result2, _ = cmd3.Output()
		s = strings.Fields(string(result2))
		//fmt.Printf("s %T,%v", s, s)
		item.name = s[0]
		item.itemType = s[1]
		item.lable = s[3]
		item.value, _ = strconv.ParseFloat(s[2], 64)
		if len(s) == 5 {
			item.lablename = s[4]
		}else{
			item.lablename = ""
		}
		fmt.Printf("\nitem is %T, %v\n", item, item)
		fmt.Printf("\nlablename is %T, %v\n", item.lablename, item.lablename)
		g = l + 1
		itemList[g] = item
	}
	return itemList
}

//监控项
func itemfor(name string) (itemvalue float64){
	var cmd *exec.Cmd
	allpath := path + "/scripts/" + name
	cmd = exec.Command("/bin/bash", allpath)
	//result, _ := cmd.Output()
	//fmt.Printf("scripts output:  %v", string(result))
	_, err := cmd.Output()
	if err != nil {
		fmt.Printf("\n执行脚本报错\n 错误=%v", err)
	}
	promPATH := path + "/prom/" + name + ".prom"
	//fmt.Println(promPATH)
	var cmd3 *exec.Cmd
	cmd3 = exec.Command("/bin/cat", promPATH)
	result2, _ := cmd3.Output()
	s := strings.Fields(string(result2))
	itemvalue,_ = strconv.ParseFloat(s[2], 64)
	return  itemvalue
}

// 监控项
func varItems(itemList  map[int]itemData)  () {
	for _, v := range itemList{
		switch  v.itemType {
		case "counter" :
			fmt.Println("暂不支持这种类型")
		case "gauge":
			if len(v.lablename) != 0 {
				var x = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: v.name,
					Help: v.name,
				},
					[]string{v.lable},
				)
				prometheus.MustRegister(x)
				fmt.Println(v.name)
				go func(name, lable, lablename string) {
					for {
						x.With(prometheus.Labels{lable: lablename}).Set(itemfor(name))
						time.Sleep(200 * time.Millisecond)
					}
				}(v.name, v.lable, v.lablename)
			}else{
				var x = prometheus.NewGauge(prometheus.GaugeOpts{
									Name: v.name,
									Help: v.name,
				})
				prometheus.MustRegister(x)
				fmt.Println(v.name)
				go func(name string){
					for {
						x.Set(itemfor(name))
						time.Sleep(200 * time.Millisecond)
						}
				}(v.name)
			}

		case "histogram":
			fmt.Println("暂不支持这种类型")
		case "summary":
			fmt.Println("暂不支持这种类型")
		}

	}

}




var addr = flag.String("listen-address", ":9000", "The address to listen on for HTTP requests.")
var path string
func main() {
	flag.StringVar(&path, "f", "filename", "指定脚本文件路径" )
	flag.Parse()
	varItems(allOutput())
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}





