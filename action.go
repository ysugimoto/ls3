package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nsf/termbox-go"
)

var mimeTypeList = map[string]int{
	"text/plain":             1,
	"text/html":              1,
	"text/css":               1,
	"application/javascript": 1,
}

type Action struct {
	object *s3.GetObjectOutput
	name   string
	offset int
}

func NewAction(object *s3.GetObjectOutput, objectName string, offset int) *Action {
	return &Action{
		object: object,
		name:   objectName,
		offset: offset,
	}
}

func (a *Action) Do() (bool, error) {
	pointer := a.displayObjectInfo()
	var act ObjectAction
	for {
		var err error
		act, err = a.chooseAction(pointer)
		if err != nil {
			continue
		}
		break
	}
	switch act {
	case Download:
		return a.doDownload()
	case View:
		return a.doView()
	case Back, None:
	}
	return false, nil
}

func (a *Action) displayObjectInfo() int {
	pointer := a.offset
	infoList := []string{
		fmt.Sprint(strings.Repeat("=", 60)),
		fmt.Sprintf("%-16s: %s\n", "Content Type", *a.object.ContentType),
		fmt.Sprintf("%-16s: %d\n", "File Size", *a.object.ContentLength),
		fmt.Sprintf("%-16s: %s\n", "Last Modified", utcToJst(*a.object.LastModified)),
		"",
		"Choose Action for this file:",
	}
	for _, info := range infoList {
		for i, r := range []rune(info) {
			termbox.SetCell(i, pointer, r, termbox.ColorDefault, termbox.ColorDefault)
		}
		pointer++
	}
	return pointer
}

func (a *Action) chooseAction(pointer int) (ObjectAction, error) {
	back := ActionCommand{op: Back, name: "Back To List"}
	view := ActionCommand{op: View, name: "View file content"}
	download := ActionCommand{op: Download, name: "Download this file"}

	actions := []ActionCommand{back}
	selects := Selectable{back}

	if _, ok := mimeTypeList[*a.object.ContentType]; ok {
		actions = append(actions, view)
		selects = append(selects, view)
	}
	actions = append(actions, download)
	selects = append(selects, download)

	selector := NewSelector(pointer)
	action, err := selector.Choose(selects)
	if err != nil {
		return None, err
	}
	switch action {
	case 0:
		return Back, nil
	case 1:
		if _, ok := mimeTypeList[*a.object.ContentType]; ok {
			return View, nil
		}
		return Download, nil
	case 2:
		return Download, nil
	default:
		return None, nil
	}
}

func (a *Action) doDownload() (bool, error) {
	width, height := termbox.Size()
	state := []rune(fmt.Sprintf("Downloading %s ...", a.name))
	for i, r := range state {
		termbox.SetCell(i, height, r, termbox.ColorWhite, termbox.ColorCyan)
	}
	for i := len(state); i < width; i++ {
		termbox.SetCell(i, height, rune(' '), termbox.ColorDefault, termbox.ColorCyan)
	}

	buffer, err := ioutil.ReadAll(a.object.Body)
	if err != nil {
		return false, err
	}
	cwd, _ := os.Getwd()
	writePath := fmt.Sprintf("%s/%s", cwd, a.name)
	if err := ioutil.WriteFile(writePath, buffer, 0644); err != nil {
		return false, err
	}
LOOP:
	for {
		select {
		case <-time.After(2 * time.Second):
			break LOOP
		}
	}
	return true, nil
}

func (a *Action) doView() (bool, error) {
	return true, nil
}
