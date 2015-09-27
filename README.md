## shweb
Write web server use shell.

## Why I doing this.
Shell is very simple to do daily work.

I think it is also possible to use shell write simple simple server for immergency use or experimental use.

## How to use

	$ go get -v github.com/codeskyblue/shweb
	$ cat > sample.cfg <<EOF
	GET / index.sh
	GET /index index.sh
	EOF

	$ cat > index.sh <<EOF
	#!/bin/bash
	echo hello world
	EOF

	$ shweb -f sample.cfg -port 4000

## More about configfile format
```
GET / index.html
GET /js	hello.js
GET /shell hello.sh
GET /doc   README.md
GET /jade  hello.jade
POST /post post.sh json
```

## CONTRIBUTE
Need your ideas, make a issue and let me known.

## LICENSE
[MIT](LICENSE)
