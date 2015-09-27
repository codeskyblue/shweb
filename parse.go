package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Unknwon/macaron"
	"github.com/go-floki/jade"
	"github.com/kballard/go-shellquote"
	"github.com/microcosm-cc/bluemonday"
	"github.com/qiniu/log"
	"github.com/russross/blackfriday"
)

func detectCType(path string) string {
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype != "" {
		// go return charset=utf8 even if the charset is not utf8
		idx := strings.Index(ctype, "; ")
		if idx > 0 {
			// remove charset; anyway, browsers are very good at guessing it.
			ctype = ctype[0:idx]
		}
	}
	return ctype
}

func parseCfgFile(file string, m *macaron.Macaron) error {
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
		if !(len(fields) >= 3 && len(fields) <= 4) {
			continue
		}

		method, patten, script := fields[0], fields[1], fields[2]
		contentType := ""
		if len(fields) == 4 {
			contentType = fields[3]
		} else {
		}
		addRoute(m, method, patten, script, contentType)
	}
	return nil
}

func addRoute(m *macaron.Macaron, method, patten, script, contentType string) {
	log.Println("Add Route:", method, patten, script)
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		m.Get(patten, NewScriptHandler(script, contentType))
	case "POST":
		m.Post(patten, NewScriptHandler(script, contentType))
	}
}

func runShellScript(ctx *macaron.Context, script string) ([]byte, error) {
	ctx.Req.ParseForm()
	//log.Println(ctx.Req.Form)
	envs := map[string]string{
		"REQ_PATH":   ctx.Req.URL.Path,
		"REQ_URI":    ctx.Req.RequestURI,
		"REQ_METHOD": ctx.Req.Method,
	}
	for key, vals := range ctx.Req.Form {
		log.Debug("Form value:", key, vals)
		envs["FORM_"+key] = vals[0]
	}
	for key, vals := range ctx.Req.PostForm {
		log.Debug("Form value:", key, vals)
		envs["POST_FORM_"+key] = vals[0]
	}
	environ := os.Environ()
	for key, val := range envs {
		environ = append(environ, key+"="+val)
	}
	cmd := exec.Command("/bin/bash", script) //.Output()
	cmd.Env = environ
	return cmd.Output()
}

func render(ctx *macaron.Context, output []byte, err error, contentType string) {
	if err != nil {
		ctx.Error(500, err.Error()+"\n"+string(output))
		return
	}
	if contentType == "" {
		contentType = "text"
	}
	switch contentType {
	case "html":
		ctx.Header().Set("Content-Type", "text/html")
	case "json":
		ctx.Header().Set("Content-Type", "application/json; charset=UTF-8")
	case "text":
		ctx.Header().Set("Content-Type", "text/plain")
	default:
		ctx.Header().Set("Content-Type", contentType)
	}
	ctx.Write(output)
}

func NewScriptHandler(script string, contentType string) func(*macaron.Context) {
	ext := filepath.Ext(script)
	return func(ctx *macaron.Context) {
		var err error
		var output []byte
		errquit := func(err error) {
			ctx.Error(500, err.Error()+"\n"+string(output))
			return
		}
		switch ext {
		case ".sh":
			output, err = runShellScript(ctx, script)
			render(ctx, output, err, contentType)
		case ".html":
			output, err = ioutil.ReadFile(script)
			render(ctx, output, err, contentType)
		case ".md":
			output, err = ioutil.ReadFile(script)
			if err != nil {
				errquit(err)
				return
			}
			unsafe := blackfriday.MarkdownCommon(output)
			html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
			render(ctx, html, nil, "html")
		case ".jade":
			tmpl, err := jade.CompileFile(script, jade.Options{})
			if err != nil {
				errquit(err)
				return
			}
			buf := bytes.NewBuffer(nil)
			err = tmpl.Execute(buf, nil)
			render(ctx, buf.Bytes(), err, "html")
		default:
			log.Warn("Unknown script ext", script, ext)
			output, err = ioutil.ReadFile(script)
			render(ctx, output, err, detectCType(script))
		}
	}
}
