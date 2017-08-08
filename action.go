package main

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nsf/termbox-go"
)

// Action for S3 object
type Action struct {

	// object which returns S3 API
	object *s3.GetObjectOutput

	// Status Writer
	status *Status

	// Injected Selector
	selector *Selector

	// Object name
	name string

	// row offset for termbox
	offset int

	// Duplicate guard
	guard chan struct{}
}

// Create Action pointer
func NewAction(object *s3.GetObjectOutput, objectName string, selector *Selector, status *Status, offset int) *Action {
	return &Action{
		object:   object,
		name:     objectName,
		offset:   offset,
		selector: selector,
		status:   status,
		guard:    make(chan struct{}, 1),
	}
}

// Do action
func (a *Action) Do() (bool, error) {
	a.guard <- struct{}{}
	defer func() {
		<-a.guard
	}()

	pointer := a.displayObjectInfo()
	a.status.Message("Choose Action for this file", 0)

	switch a.chooseAction(pointer) {
	case Download:
		return a.doDownload()
	case Back, None:
		return false, nil
	default:
		return false, nil
	}
}

func (a *Action) resize() {
	if len(a.guard) > 0 {
		a.displayObjectInfo()
	}
}

// Display object info
func (a *Action) displayObjectInfo() (pointer int) {
	pointer = a.offset
	infoList := [6]string{
		"",
		fmt.Sprint(strings.Repeat("=", 60)),
		fmt.Sprintf("%-16s: %s\n", "Content Type", *a.object.ContentType),
		fmt.Sprintf("%-16s: %d (bytes)\n", "File Size", *a.object.ContentLength),
		fmt.Sprintf("%-16s: %s\n", "Last Modified", utcToJst(*a.object.LastModified)),
		"",
	}
	for _, info := range infoList {
		for i, r := range []rune(info) {
			termbox.SetCell(i, pointer, r, termbox.ColorDefault, termbox.ColorDefault)
		}
		pointer++
	}
	return
}

// Choose action for selected object
func (a *Action) chooseAction(pointer int) ObjectAction {
	back := ActionCommand{op: Back, name: "Back To List"}
	download := ActionCommand{op: Download, name: "Download this file"}

	actions := ActionList{back, download}

	a.selector.SetOffset(pointer).WithOutFilter()
	defer func() {
		a.selector.SetOffset(a.offset).WithFilter()
	}()

	action, _ := a.selector.Choose(actions.Selectable())
	switch action {
	case 0:
		return Back
	case 2:
		return Download
	default:
		return None
	}
}

// Download object to current working directory
func (a *Action) doDownload() (bool, error) {
	a.status.Info(fmt.Sprintf("Downloading %s ...", a.name), 0)

	buffer, err := ioutil.ReadAll(a.object.Body)
	if err != nil {
		return false, err
	}
	cwd, _ := os.Getwd()
	writePath := fmt.Sprintf("%s/%s", cwd, a.name)
	if err := ioutil.WriteFile(writePath, buffer, 0644); err != nil {
		<-a.status.Error("Failed to download", 1)
		return false, err
	}
	go func() {
		<-a.status.Info("Downloaded completely!", 1)
	}()
	return false, nil
}
