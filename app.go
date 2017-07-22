package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nsf/termbox-go"
	"strings"
)

type App struct {
	bucket  string
	prefix  []string
	object  string
	service *s3.S3
}

func NewApp(service *s3.S3, bucket string) (*App, error) {
	if err := termbox.Init(); err != nil {
		return nil, err
	}
	return &App{
		service: service,
		bucket:  bucket,
		prefix:  []string{},
	}, nil
}

func (a *App) Terminate() {
	termbox.Close()
}

func (a *App) Run() error {
	if a.bucket == "" {
		if err := a.chooseBuckets(); err != nil {
			return err
		}
	}
	if err := a.chooseObject(); err != nil {
		return err
	}
	return nil
}

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
	location := fmt.Sprintf("s3://%s%s%s", b, p, o)
	for i, r := range []rune(location) {
		termbox.SetCell(i, 0, r, termbox.ColorGreen, termbox.ColorDefault)
	}
	termbox.Flush()
}

// Choose bucket from list
func (a *App) chooseBuckets() error {
	result, err := a.service.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}
	buckets := []*Bucket{}
	selects := Selectable{}
	for _, b := range result.Buckets {
		bucket := NewBucket(b)
		buckets = append(buckets, bucket)
		selects = append(selects, bucket)
	}
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	a.writeHeader()

	selector := NewSelector(1)
	index, err := selector.Choose(selects)
	if err != nil {
		return err
	}
	a.bucket = buckets[index].name
	return nil
}

// Choose from object list
func (a *App) chooseObject() error {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(a.bucket),
	}
	if len(a.prefix) > 0 {
		input = input.SetPrefix(strings.Join(a.prefix, "/") + "/")
	}
	result, err := a.service.ListObjects(input)
	if err != nil {
		return err
	}
	parent := NewParentObject()
	objects := []*Object{parent}
	selects := Selectable{parent}
	for _, o := range a.formatObjects(result.Contents) {
		objects = append(objects, o)
		selects = append(selects, o)
	}

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	a.writeHeader()
	selector := NewSelector(1)
	index, err := selector.Choose(selects)
	if err != nil {
		return err
	}
	selected := objects[index]
	switch {
	case selected.key == "../":
		a.object = ""
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

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	a.writeHeader()
	action := NewAction(result, a.object, 1)
	return action.Do()
}

func (a *App) formatObjects(s3Objects []*s3.Object) []*Object {
	replace := ""
	if len(a.prefix) > 0 {
		replace = strings.Join(a.prefix, "/") + "/"
	}
	objects := []*Object{}
	unique := map[string]struct{}{}
	for _, o := range s3Objects {
		isDir := false
		key := strings.Replace(*o.Key, replace, "", 1)
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
