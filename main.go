package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zzjbattlefield/IM_GO/api"
	"github.com/zzjbattlefield/IM_GO/connect"
	"github.com/zzjbattlefield/IM_GO/logic"
	"github.com/zzjbattlefield/IM_GO/task"
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
	case "connect":
		connect.New().Run()
	case "task":
		task.New().Run()
	default:
		fmt.Println("param error!")
		return
	}
	fmt.Println(fmt.Printf("success run module %s", module))
	fmt.Println(fmt.Sprintf("run %s module done!", module))
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	fmt.Println("Server exiting")
}
