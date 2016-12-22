This editor mimics a single window of [acme(1)](http://www.unix.com/man-page/plan9/1/acme/).
The author uses this editor on a daily basis, but there are many missing features and bugs.

Installation and usage:
```
go get -u -t sigint.ca/graphics/cmd/edit
edit [file]
```

The font can be set by setting the FONT environment variable:
```
FONT=/Library/Fonts/Comic\ Sans\ MS.ttf edit
```

The author uses a wrapper script to background the process:
```
$ cat ~/bin/bgedit 
#!/bin/sh
edit </dev/stdin >/dev/stdout 2>/dev/stderr $@ & disown
```

Some supported acme features:
- Mouse button chording
- Double click selection rules
- Right click to search
- B2 | (pipe) commands (e.g. |sort)
- History (some bugs lurking here)
- Auto indentation
- Acme style B1/B2/B3 scrollbar behaviour
- More

Features not in acme or differing from acme:
- Cross platform (though currently only tested on OS X)
- Smooth scrolling
- TTF fonts
- Click to focus tag or editor
- C-S to save, C-A to select all
- B2 click of a shell command launches a new editor containing output
- More

Currently unsupported:
- B2 < and > commands
- Plumber
- 9P filesystem
- Window management
- Many standard acme commands
- Directory browser
- More
