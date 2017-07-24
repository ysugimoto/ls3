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
	status  *Status
}

func NewApp(service *s3.S3, bucket string) (*App, error) {
	if err := termbox.Init(); err != nil {
		return nil, err
	}
	return &App{
		service: service,
		bucket:  bucket,
		prefix:  []string{},
		status:  NewStatus(1),
	}, nil
}

func (a *App) Terminate() {
	termbox.Close()
}

func (a *App) Run() error {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	a.writeHeader()
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
	location := fmt.Sprintf("Location: s3://%s%s%s", b, p, o)
	for i, r := range []rune(location) {
		termbox.SetCell(i, 0, r, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
	}
	termbox.Flush()
}

// Choose bucket from list
func (a *App) chooseBuckets() error {
	a.status.Message("Retrive bucket list...", 0)
	result, err := a.service.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}
	buckets := Buckets{}
	for _, b := range result.Buckets {
		buckets = append(buckets, NewBucket(b))
	}
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	a.writeHeader()

	a.status.Message("Choose bucket", 0)
	selector := NewSelector(2)
	index, err := selector.Choose(buckets.Selectable())
	if err != nil {
		a.status.Clear()
		return err
	}
	a.status.Clear()
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
	a.status.Message("Retrive object list...", 0)
	result, err := a.service.ListObjects(input)
	if err != nil {
		return err
	}
	objects := Objects{NewParentObject()}
	for _, o := range a.formatObjects(result.Contents) {
		objects = append(objects, o)
	}

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	a.writeHeader()

	a.status.Message("Choose object", 0)
	selector := NewSelector(2)
	index, err := selector.Choose(objects.Selectable())
	if err != nil {
		a.status.Clear()
		return err
	}
	a.status.Clear()
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
	action := NewAction(result, a.object, 2)
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
