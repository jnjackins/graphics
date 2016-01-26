package main

import "strings"

var tagline string

func loadTag() {
	var newTags []string
	if filename != "" {
		newTags = append(newTags, filename)
	}
	newTags = append(newTags, "Undo", "Redo")
	if filename != "" {
		newTags = append(newTags, "Put")
	}
	newTags = append(newTags, "Exit")

	newTagline := strings.Join(newTags, " ")
	if newTagline == tagline {
		return
	}
	tagline = newTagline

	tagWidget.ed.Load([]byte(tagline))
}
