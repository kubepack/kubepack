package main

import (
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"kmodules.xyz/client-go/logs"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if err := NewCmdFuse().Execute(); err != nil {
		log.Fatalln("error:", err)
	}
}
