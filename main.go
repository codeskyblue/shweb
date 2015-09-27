/*
Config file patten
---
GET / index.html
POST /hi handle-post.sh
*/

package main

import (
	"flag"
	"os"

	"github.com/Unknwon/macaron"
	"github.com/qiniu/log"
)

var (
	srvPort = flag.Int("port", 4000, "server port")
	cfgFile = flag.String("f", "", "config file")

	m = macaron.Classic()
)

func readConfig() {
	flag.Parse()
	if !FileExists(*cfgFile) {
		log.Fatal("Need config file")
	}
	parseCfgFile(*cfgFile, m)
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func main() {
	readConfig()

	log.SetOutputLevel(log.Ldebug)
	m.Use(macaron.Renderer())
	m.Run(*srvPort)
}
