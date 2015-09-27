/*
Config file patten
---
GET / index.html
POST /hi handle-post.sh
*/

package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Unknwon/macaron"
	"github.com/kballard/go-shellquote"
	"github.com/qiniu/log"
)

var (
	srvPort = flag.Int("port", 4000, "server port")
	cfgFile = flag.String("f", "", "config file")

	m = macaron.Classic()
)

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func readConfig() {
	flag.Parse()
	if !FileExists(*cfgFile) {
		log.Fatal("Need config file")
	}
	parseCfgFile(*cfgFile)
}

func parseCfgFile(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	rd := bufio.NewReader(fd)
	for {
		bline, _, err := rd.ReadLine() // Just ignore isPrefix(maybe not good)
		if err != nil {
			break
		}
		line := string(bline)
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields, err := shellquote.Split(line)
		if err != nil {
			log.Printf("Shellquote parse error: %v", err)
			continue
		}
		if len(fields) != 3 {
			continue
		}

		method, patten, script := fields[0], fields[1], fields[2]
		addRoute(method, patten, script)
	}
	return nil
}

func runShellScript(script string) ([]byte, error) {
	return exec.Command("/bin/bash", script).Output()
}

func NewScriptHandler(script string) func(*macaron.Context) {
	ext := filepath.Ext(script)
	return func(ctx *macaron.Context) {
		var err error
		var output []byte
		defer func() {
			if err != nil {
				ctx.Error(500, err.Error()+"\n"+string(output))
			}
		}()
		switch ext {
		case ".sh":
			output, err = runShellScript(script)
			if err == nil {
				ctx.JSON(200, string(output))
			}
		default:
			log.Warn("Unknown script ext", script, ext)
			output, err = ioutil.ReadFile(script)
			if err == nil {
				ctx.Write(output)
			}
		}
	}
}

func addRoute(method, patten, script string) {
	log.Println("Add Route:", method, patten, script)
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		m.Get(patten, NewScriptHandler(script))
	}
}

func main() {
	readConfig()

	log.SetOutputLevel(log.Ldebug)
	m.Use(macaron.Renderer())
	m.Run(*srvPort)
}
