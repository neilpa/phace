package main

import (
	"html/template"
)

// TODO Fix orientation
type photoPage struct {
	ImageUrl   string
	Prev, Next string
}

var photoHtml = template.Must(template.New("photo").Parse(
	`<!doctype html>
<html>
<head>
  <style>
    html, body {
      background: black;
      height: 100%; // https://stackoverflow.com/a/18745921
      margin: 0;
      padding: 0;
    }
    img { // https://stackoverflow.com/a/30794589
      width: 100%;
      height: 100%;
      object-fit: contain;
    }
  </style>
</head>
<body>
  <img style='width:100%;height:100%;object-fit:contain' src='{{ .ImageUrl }}'/>
</body>
<script>
  document.addEventListener('keydown', function(e) {
    // Ignore IME events: https://developer.mozilla.org/en-US/docs/Web/API/Document/keydown_event
    if (e.isComposing || e.keyCode === 229) {
      return;
    }
    switch (e.keyCode) {
      case 37: // left
        window.location = '{{ .Prev }}'
        break;
      case 38: // up
        break;
      case 39: // right
        window.location = '{{ .Next }}'
        break;
      case 40: // down
        break;
    }
  })
</script>
</html>`))
