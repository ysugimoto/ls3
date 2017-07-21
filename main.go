package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ysugimoto/pecolify"
)

// CLI option value struct
type CLI struct {
	// Initial bucket name
	bucket string

	// Using profile name
	profile string

	// Determine region
	region string

	// Using profile from environment
	env bool

	// Show help
	help bool
}

var cli CLI = CLI{}
var service *s3.S3
var p = pecolify.New()
var mimeTypeList = map[string]struct{}{
	"text/plain":             struct{}{},
	"text/html":              struct{}{},
	"text/css":               struct{}{},
	"application/javascript": struct{}{},
}

// init() for parsing command line args
func init() {
	flag.StringVar(&cli.bucket, "bucket", "", "Using bucket name")
	flag.StringVar(&cli.profile, "profile", "", "Use profile name")
	flag.BoolVar(&cli.env, "env", false, "Use credentials from environment")
	flag.StringVar(&cli.region, "region", "", "region name")
	flag.BoolVar(&cli.help, "help", false, "show usage")
	flag.Parse()

	if cli.help {
		showUsage()
		os.Exit(0)
	}
}

// show usage
func showUsage() {
	help := `========================================================================
ls3 : AWS S3 file explorer on CLI
========================================================================
Usage:
  ls3 [options]

Options:
  -profile [profile name] : Use profile name which is written in ~/.aws/credentials
                            If not supplied, use default profile
  -env                    : Use credentials from environment variable
                            You need to export AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
  -bucket                 : Initial bucket name
  -region [region name]   : Determine region (default: ap-northeast-1)
  -help                   : Show this help
`
	fmt.Println(help)
}

// Main function
func main() {
	sess := session.Must(session.NewSession())
	var conf *aws.Config
	if cli.env {
		conf = configFromEnv(cli.region)
	} else {
		conf = configFromProfile(cli.profile, cli.region)
	}

	service = s3.New(sess, conf)
	var bucket string
	var err error
	if cli.bucket != "" {
		bucket = cli.bucket
	} else {
		bucket, err = chooseBuckets()
		if err != nil {
			return
		}
	}
	fmt.Printf("Retriving %s bucket...\n", bucket)
	chooseObjects(bucket, []string{})
}

// Create aws.Config from profile
func configFromProfile(profileName, region string) *aws.Config {
	if region == "" {
		region = "ap-northeast-1"
	}
	provider := &credentials.SharedCredentialsProvider{
		Profile: profileName,
	}
	creds := credentials.NewCredentials(provider)
	return aws.NewConfig().
		WithCredentials(creds).
		WithRegion(region)
}

// Create aws.Config from environment
func configFromEnv(region string) *aws.Config {
	if region == "" {
		region = "ap-northeast-1"
	}
	return aws.NewConfig().
		WithCredentials(credentials.NewEnvCredentials()).
		WithRegion(region)
}

// Choose bucket from list
func chooseBuckets() (string, error) {
	result, err := service.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		panic(err)
	}
	buckets := []string{}
	for i := 0; i < len(result.Buckets); i++ {
		name := *result.Buckets[i].Name
		buckets = append(buckets, fmt.Sprintf("[Bucket] %s", name))
	}

	selected, err := p.Transform(buckets)
	if err != nil {
		return "", err
	}
	return selected[9:], nil
}

// Choose from object list
func chooseObjects(bucket string, prefix []string) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}
	if len(prefix) > 0 {
		input = input.SetPrefix(strings.Join(prefix, "/") + "/")
	}
	result, err := service.ListObjects(input)
	if err != nil {
		fmt.Println(err)
	}

	selected, err := p.Transform(formatObjectList(result.Contents, prefix))
	if err != nil {
		return
	}

	if selected == "../" {
		if len(prefix) > 0 {
			prefix = prefix[0 : len(prefix)-1]
		} else {
			b, err := chooseBuckets()
			if err != nil {
				return
			}
			bucket = b
			prefix = []string{}
		}
	} else if strings.HasSuffix(selected, "/") {
		selected = strings.TrimRight(selected[32:], "/")
		prefix = append(prefix, selected)
	} else {
		selected = strings.TrimRight(selected[32:], "/")
		if isEnd := displayObjectActions(bucket, prefix, selected); isEnd {
			return
		}
	}
	chooseObjects(bucket, prefix)
}

// Display object info for slected object
func displayObjectActions(bucket string, prefix []string, object string) bool {
	dir := ""
	if len(prefix) > 0 {
		dir = strings.Join(prefix, "/") + "/"
	}

	result, err := service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("%s%s", dir, object)),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("%-16s: s3://%s/%s%s\n", "S3 Location", bucket, dir, object)
	fmt.Printf("%-16s: %s\n", "Content Type", *result.ContentType)
	fmt.Printf("%-16s: %d\n", "File Size", *result.ContentLength)
	fmt.Printf("%-16s: %s\n", "Last Modified", utcToJst(*result.LastModified))
	fmt.Println("")
	return objectAction(result, object)
}

// Display object info and select action
func objectAction(result *s3.GetObjectOutput, object string) bool {
	var cmd string
	line := "Actions? [d:Download"
	ct := *result.ContentType
	_, isTextMime := mimeTypeList[ct]
	if isTextMime {
		line += ", v:View"
	}
	line = line + ", b:Back to list]: "
	fmt.Print(line)
	fmt.Scanf("%s", &cmd)

	switch cmd {
	case "d":
		fmt.Printf("Downloading %s...", object)
		buf, err := ioutil.ReadAll(result.Body)
		if err != nil {
			fmt.Println(err)
			return false
		}
		var cwd string
		cwd, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
			return false
		}
		if err := ioutil.WriteFile(fmt.Sprintf("%s/%s", cwd, object), buf, 0644); err != nil {
			fmt.Println(err)
			return true
		}
		fmt.Println("Download completed")
		return false
	case "v":
		if isTextMime {
			if buf, err := ioutil.ReadAll(result.Body); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(string(buf))
			}
			return false
		} else {
			fmt.Println("Invalid access v action")
		}
		return true
	case "b":
		return false
	default:
		fmt.Println("Invalid command")
		return objectAction(result, object)
	}
}

// Fromat object list from S3 response
func formatObjectList(result []*s3.Object, prefix []string) (objects []string) {
	var rep string
	if len(prefix) > 0 {
		rep = strings.Join(prefix, "/") + "/"
	}
	objects = append(objects, "../")
	unique := map[string]struct{}{}
	for i := 0; i < len(result); i++ {
		isDir := false
		key := *result[i].Key
		lastModified := *result[i].LastModified
		size := *result[i].Size
		key = strings.Replace(key, rep, "", 1)
		if strings.Contains(key, "/") {
			spl := strings.Split(key, "/")
			if _, exists := unique[spl[0]]; exists {
				continue
			}
			unique[spl[0]] = struct{}{}
			key = spl[0] + "/"
			isDir = true
		}
		jst := utcToJst(lastModified)
		if isDir {
			objects = append(objects, fmt.Sprintf("%s %10s  %s", jst, "-", key))
		} else {
			objects = append(objects, fmt.Sprintf("%s %10d  %s", jst, size, key))
		}
	}
	return
}

// We're living in Asia/Tokyo location :)
var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

// Transform from UTC to JST
func utcToJst(utc time.Time) string {
	jst := utc.In(JST)
	return jst.Format("2006-01-02 15:03:04")
}
