package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"time"
)

// We're living in Asia/Tokyo location :)
var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

// Transform from UTC to JST
func utcToJst(utc time.Time) string {
	jst := utc.In(JST)
	return jst.Format("2006-01-02 15:03:04")
}

type Bucket struct {
	name string
}

func NewBucket(b *s3.Bucket) *Bucket {
	return &Bucket{
		name: *b.Name,
	}
}

func (b *Bucket) String() string {
	return fmt.Sprintf("[Bucket] %s", b.name)
}

type Object struct {
	size         int64
	key          string
	lastModified time.Time
	dir          bool
	parent       bool
}

func NewObject(key string, size int64, lastModified time.Time, dir bool) *Object {
	return &Object{
		size:         size,
		key:          key,
		lastModified: lastModified,
		dir:          dir,
	}
}

func NewParentObject() *Object {
	return &Object{
		key:    "../",
		parent: true,
	}
}
func (o *Object) String() string {
	if o.parent {
		return o.key
	} else if o.dir {
		return fmt.Sprintf("%s %10s  %s/", utcToJst(o.lastModified), "-", o.key)
	} else {
		return fmt.Sprintf("%s %10d  %s", utcToJst(o.lastModified), o.size, o.key)
	}
}

type ObjectAction int

const (
	Back ObjectAction = iota
	View
	Download
	None = 999
)

type ActionCommand struct {
	op   ObjectAction
	name string
}

func (a ActionCommand) String() string {
	return a.name
}
