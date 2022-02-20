package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"time"
)

var countFiles int32 = 0
var countFiles1 int32 = 0

var s *string

const maxWorks = int32(32)

var (
	res      = make(chan bool)
	path     = make(chan string)
	workDone = make(chan bool)
	works    = maxWorks
)

func init() {
	s = flag.String("f", ".", "use for path")
}
func main() {
	flag.Parse()
	start := time.Now()
	works--
	go multiFindFile(*s, true)
	deal()
	first := time.Now()
	fmt.Println("multiFind", first.Sub(start).Milliseconds())
	start = time.Now()
	findFile(*s)
	fmt.Println("findFiles", time.Now().Sub(start).Milliseconds())
	fmt.Println(countFiles1, countFiles)
	fmt.Println(works)
}

func deal() {
	for true {
		select {
		case <-res:
			countFiles++
		case path_ := <-path:
			go multiFindFile(path_, true)

		case <-workDone:
			if atomic.AddInt32(&works, 1) == maxWorks {
				return
			}
		}
	}

}

func multiFindFile(path_ string, isMaster bool) {
	/*
		isMaster: true代表是创建了新的goroutine，否则是在当前goroutine递归
	*/
	if infos, err := ioutil.ReadDir(path_); err == nil {
		for _, file := range infos {
			if file.IsDir() {
				if atomic.AddInt32(&works, -1) < 0 {
					atomic.AddInt32(&works, 1)
					multiFindFile(path_+"/"+file.Name()+"/", false)
					continue
				}
				path <- path_ + "/" + file.Name() + "/"
				continue
			}
			res <- true
		}
		if isMaster {
			workDone <- true
		}
		return
	}
	// 避免权限不够造成死锁
	if isMaster {
		workDone <- true
	}
}

func findFile(path string) {
	if infos, err := ioutil.ReadDir(path); err == nil {
		for _, file := range infos {
			if file.IsDir() {
				findFile(path + "/" + file.Name() + "/")
				continue
			}
			countFiles1++
		}
	}
}
