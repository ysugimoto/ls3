package main

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nsf/termbox-go"
)

// Enable to view file content MimeTypes
var mimeTypeList = map[string]int{
	"text/plain":             1,
	"text/html":              1,
	"text/css":               1,
	"application/javascript": 1,
}

// Action for S3 object
type Action struct {

	// object which returns S3 API
	object *s3.GetObjectOutput

	// Status Writer
	status *Status

	// Object name
	name string

	// row offset for termbox
	offset int
}

// Create Action pointer
func NewAction(object *s3.GetObjectOutput, objectName string, offset int) *Action {
	return &Action{
		object: object,
		name:   objectName,
		offset: offset,
		status: NewStatus(1),
	}
}

// Do action
func (a *Action) Do() (bool, error) {
	pointer := a.displayObjectInfo()
	a.status.Message("Choose Action for this file", 0)
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
		break
		// return a.doView()
	case Back, None:
	}
	return false, nil
}

// Display object info
func (a *Action) displayObjectInfo() (pointer int) {
	pointer = a.offset
	infoList := []string{
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
func (a *Action) chooseAction(pointer int) (ObjectAction, error) {
	back := ActionCommand{op: Back, name: "Back To List"}
	// view := ActionCommand{op: View, name: "View file content"}
	download := ActionCommand{op: Download, name: "Download this file"}

	actions := ActionList{back}
	// if _, ok := mimeTypeList[*a.object.ContentType]; ok {
	// 	actions = append(actions, view)
	// }
	actions = append(actions, download)

	selector := NewSelector(pointer).WithOutFilter()
	action, err := selector.Choose(actions.Selectable())
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
	<-a.status.Info("Downloaded completely!", 2)
	return false, nil
}

// Currently not implemented
// func (a *Action) doView() (bool, error) {
// 	return true, nil
// }
