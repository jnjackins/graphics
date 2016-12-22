Graphical tools and packages in Go.

[editor](http://godoc.org/sigint.ca/graphics/editor) provides a
graphical, editable text area widget.

cmd/edit is an almost-pure-Go text editor which uses the editor
package, and is intended to resemble
[acme](http://9p.io/magic/man2html/1/acme) without the window
management — for now, at least. The non-pure-Go part is the
[shiny](https://godoc.org/golang.org/x/exp/shiny) driver, though
in theory a pure-Go shiny driver could be developed (e.g. using
Plan 9's /dev/draw, or Linux's frame buffer device).

![screenshot](https://cloud.githubusercontent.com/assets/449232/21436229/aa7cb0a6-c83a-11e6-8319-63dbbf23935c.png)
