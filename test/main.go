package main

import (
	"sort"
	"strings"

	"github.com/astaxie/beego"
)

func varsToString(vars map[string]string) string {
	var slice []string
	for _, v := range vars {
		slice = append(slice, v)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
	return strings.Join(slice, ",")
}

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {
	this.Ctx.WriteString("hello world")
}

func main() {
	beego.Router("/", &MainController{})
	beego.Run(":9098")
}
