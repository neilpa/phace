package main

import (
	"html/template"

	"neilpa.me/phace"
)

type photoPage struct {
	URL string
	Photo *phace.Photo
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
	  overflow: hidden; // avoid vertical scrollbar on firefox
    }
    body {
	  position: relative;
    }
    img, div { // https://stackoverflow.com/a/30794589
      width: 100%;
      height: 100%;
      object-fit: contain;
      // TODO sadly only works on firefox
	  //image-orientation: from-image;
	  position: absolute;
    }
    // TODO these end up cropping the image depending on rotation
    // may resort to rotating the image on the service
    //img.orient2 {
    //  transform: rotateY(180deg);
    //}
    //img.orient3 {
    //  transform: rotate(180deg);
    //}
    //img.orient4 {
    //  transform: 'rotate(180deg) rotateY(180deg);
    //}
    //img.orient5 {
    //  transform: rotate(270deg) rotateY(180deg);
    //}
    //img.orient6 {
    //  transform: rotate(90deg);
    //}
    //img.orient7 {
    //  transform: rotate(90deg) rotateY(180deg);
    //}
    //img.orient8 {
    //  transform: rotate(270deg);
    //}
  </style>
</head>
<body>
  <img class='orient{{ .Photo.Orientation }}' src='/{{ .Photo.ImagePath }}'/>
  <img class='orient{{ .Photo.Orientation }}' src='/overlay/{{ .Photo.UUID }}'/>
  <div>some text</div>
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
