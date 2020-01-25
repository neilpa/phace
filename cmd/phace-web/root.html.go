package main

import (
	"html/template"
)

var rootHtml = template.Must(template.New("root").Parse(
`<!doctype html>
<html>
<body>
  <ul>{{ range . }}
    <li><a href='{{ .URL }}'>{{ .Photo.ImageDate }}</a></li>
  {{ end }}</ul>
</body>
</html>`))
