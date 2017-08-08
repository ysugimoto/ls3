package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nsf/termbox-go"
	"strings"
)

// Terminal application struct
type App struct {

	// Bucket name
	bucket string

	// Object prefixes
	prefix []string

	// Selected object name
	object string

	// S3 service instance
	service *s3.S3

	// Status writer
	status *Status

	// Selector instance
	selector *Selector

	// Action instance
	action *Action
}

// Create new application
func NewApp(service *s3.S3, bucket string) (*App, error) {
	if err := termbox.Init(); err != nil {
		return nil, err
	}
	app := &App{
		service: service,
		bucket:  bucket,
		prefix:  []string{},
	}
	app.status = NewStatus(1)
	app.selector = NewSelector(2, app.status)
	return app, nil
}

// Terminate application
func (a *App) Terminate() {
	termbox.Close()
}

// Clear the termbox
func (a *App) Clear() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

// Run application
func (a *App) Run() error {
	a.Clear()
	a.writeHeader()

	// Observe termbox events in goroutine
	stop := make(chan struct{}, 1)
	go a.eventLoop(stop)

	if a.bucket == "" {
		if err := a.chooseBuckets(); err != nil {
			stop <- struct{}{}
			return err
		}
	}
	if err := a.chooseObject(); err != nil {
		stop <- struct{}{}
		return err
	}

	// successfully ended application
	stop <- struct{}{}
	return nil
}

func (a *App) eventLoop(stop chan struct{}) {
	queue := make(chan termbox.Event, 1)
	go func() {
		for {
			evt := <-queue
			switch evt.Type {
			case termbox.EventKey:
				logger.log("termbox keyEvent handled")
				// Send key event to selector
				a.selector.keyPress(evt)
			case termbox.EventResize:
				logger.log("termbox resizeEvent handled")
				a.Clear()
				a.writeHeader()
				a.status.resize(evt.Width, evt.Height)
				if a.action != nil {
					a.action.resize()
				}
				a.selector.resize(evt.Width, evt.Height)
			}
		}
	}()
	for {
		select {
		case <-stop:
			return
		default:
			queue <- termbox.PollEvent()
		}
	}
}

// Write application header
func (a *App) writeHeader() {
	var b, p, o string
	if a.bucket != "" {
		b = a.bucket + "/"
	}
	if len(a.prefix) > 0 {
		p = strings.Join(a.prefix, "/") + "/"
	}
	if a.object != "" {
		o = a.object
	}
	location := fmt.Sprintf("Location: s3://%s%s%s", b, p, o)
	for i, r := range []rune(location) {
		termbox.SetCell(i, 0, r, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
	}
}

// Choose bucket from list
func (a *App) chooseBuckets() error {
	a.status.Message("Retriving bucket list...", 0)
	result, err := a.service.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}
	buckets := Buckets{}
	for _, b := range result.Buckets {
		buckets = append(buckets, NewBucket(b))
	}
	a.Clear()
	a.writeHeader()

	a.status.Message("Choose bucket", 0)
	index, err := a.selector.Choose(buckets.Selectable())
	if err != nil {
		a.status.Clear()
	} else {
		a.status.Clear()
		a.bucket = buckets[index].name
	}
	return err
}

// Choose from object list
func (a *App) chooseObject() error {
	a.object = ""
	input := &s3.ListObjectsInput{
		Bucket: aws.String(a.bucket),
	}
	if len(a.prefix) > 0 {
		input = input.SetPrefix(strings.Join(a.prefix, "/") + "/")
	}
	a.status.Message("Retriving object list...", 0)
	result, err := a.service.ListObjects(input)
	if err != nil {
		return err
	}
	objects := Objects{NewParentObject()}
	for _, o := range formatObjects(result.Contents, a.prefix) {
		objects = append(objects, o)
	}

	a.Clear()
	a.writeHeader()

	a.status.Message("Choose object", 0)
	index, err := a.selector.Choose(objects.Selectable())
	if err != nil {
		a.status.Clear()
		return err
	}

	a.status.Clear()
	selected := objects[index]
	switch {
	case selected.key == "../": // selcted parent directory
		a.object = ""
		// if prefix is empty, back to choose bucket
		if len(a.prefix) == 0 {
			a.bucket = ""
			if err := a.chooseBuckets(); err != nil {
				return err
			}
		} else {
			a.prefix = a.prefix[0 : len(a.prefix)-1]
		}
	case selected.dir:
		a.object = ""
		a.prefix = append(a.prefix, selected.key)
		logger.log("Directory selected" + selected.key)
	default:
		a.object = selected.key
		if isEnd, err := a.objectAction(); err != nil {
			return err
		} else if isEnd {
			return nil
		}
	}
	return a.chooseObject()
}

// Display action for object
func (a *App) objectAction() (bool, error) {
	dir := ""
	if len(a.prefix) > 0 {
		dir = strings.Join(a.prefix, "/") + "/"
	}

	result, err := a.service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(fmt.Sprintf("%s%s", dir, a.object)),
	})
	if err != nil {
		return true, err
	}

	a.Clear()
	a.writeHeader()
	a.action = NewAction(result, a.object, a.selector, a.status, 2)
	defer func() {
		a.action = nil
	}()
	return a.action.Do()
}

// Filter and format object list
func formatObjects(s3Objects []*s3.Object, prefix []string) Objects {
	replace := ""
	if len(prefix) > 0 {
		replace = strings.Join(prefix, "/") + "/"
	}
	objects := Objects{}
	unique := map[string]struct{}{}
	for _, o := range s3Objects {
		isDir := false
		key := strings.Replace(*o.Key, replace, "", 1)

		// If key contains "/", we deal with it as directory
		if strings.Contains(key, "/") {
			parts := strings.Split(key, "/")
			if _, exist := unique[parts[0]]; exist {
				continue
			}
			unique[parts[0]] = struct{}{}
			key = parts[0]
			isDir = true
		}
		objects = append(objects, NewObject(key, *o.Size, *o.LastModified, isDir))
	}

	return objects
}
