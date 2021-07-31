package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
)

/*

 */

func main() {

	flag.Set("logtostderr", "true")
	flag.Parse()
	defer glog.Flush()

	glog.Info("Howdy")

	fmt.Println("hello world")

}
