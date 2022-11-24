package main

import (
	"flag"
	"fmt"

	"github.com/zzjbattlefield/IM_GO/api"
	"github.com/zzjbattlefield/IM_GO/logic"
)

func main() {
	var module string
	flag.StringVar(&module, "module", "", "")
	flag.Parse()
	fmt.Println(fmt.Printf("start run module %s", module))
	switch module {
	case "api":
		api.Run()
	case "logic":
		logic.New().Run()
	default:
		fmt.Println("param error!")
	}
	fmt.Println(fmt.Printf("success run module %s", module))
}
