package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ysugimoto/pecolify"
)

const UTC_LAYOUT = "2016-01-02 15:03:04 -700 MST"

type CLI struct {
	profile string
	region  string
	env     bool
	help    bool
}

var cli CLI = CLI{}
var service *s3.S3
var p = pecolify.New()
var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

// init() for parsing command line args
func init() {
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
  -region [region name]   : Determine region (default: ap-northeast-1)
  -help                   : Show this help
`
	fmt.Println(help)
}

// Create aws.Config from profile
func ConfigFromProfile(profileName, region string) *aws.Config {
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
func ConfigFromEnv(region string) *aws.Config {
	if region == "" {
		region = "ap-northeast-1"
	}
	return aws.NewConfig().
		WithCredentials(credentials.NewEnvCredentials()).
		WithRegion(region)
}

// Main function
func main() {
	sess := session.Must(session.NewSession())
	var conf *aws.Config
	if cli.env {
		conf = ConfigFromEnv(cli.region)
	} else {
		conf = ConfigFromProfile(cli.profile, cli.region)
	}

	service = s3.New(sess, conf)
	bucket, err := chooseBuckets()
	if err != nil {
		return
	}
	fmt.Printf("Retriving %s bucket...\n", bucket)
	chooseObjects(bucket, []string{})
}

func chooseBuckets() (string, error) {
	result, err := service.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		panic(err)
	}
	buckets := []string{}
	for i := 0; i < len(result.Buckets); i++ {
		name := *result.Buckets[i].Name
		buckets = append(buckets, name)
	}

	selected, err := p.Transform(buckets)
	if err != nil {
		return "", err
	}
	return selected, nil
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
	objects := formatObjectList(result.Contents, prefix)
	selected, err := p.Transform(objects)
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
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", strings.Join(prefix, "/"), object)),
	}
	result, err := service.GetObject(input)
	if err != nil {
		panic(err)
	}
	jst, _ := utcToJst(*result.LastModified)
	fmt.Println(strings.Repeat("=", 60))
	if len(prefix) > 0 {
		fmt.Printf("%-16s: s3://%s/%s\n", "Location", bucket, object)
	} else {
		fmt.Printf("%-16s: s3://%s/%s/%s\n", "Location", bucket, strings.Join(prefix, "/"), object)
	}
	fmt.Printf("%-16s: %s\n", "ContentType", *result.ContentType)
	fmt.Printf("%-16s: %d\n", "FileSize", *result.ContentLength)
	fmt.Printf("%-16s: %s\n", "LastModified", jst)
	fmt.Println("")
	return acceptAction(result)
}

func acceptAction(result *s3.GetObjectOutput) bool {
	ct := *result.ContentType
	fmt.Print("Actions? [d:Download")
	if strings.HasPrefix(ct, "text/") || strings.HasPrefix(ct, "application/") {
		fmt.Print(", v:View")
	}
	fmt.Print(", b:Back to list]: ")
	var cmd string
	fmt.Scanf("%s", &cmd)
	switch cmd {
	case "d":
		fmt.Println("Download")
		return true
	case "v":
		if strings.HasPrefix(ct, "text/") || strings.HasPrefix(ct, "application/") {
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
		return acceptAction(result)
	}
}

// Fromat object list from S3 response
func formatObjectList(result []*s3.Object, prefix []string) (objects []string) {
	objects = append(objects, "../")
	unique := map[string]struct{}{}
	var rep string
	if len(prefix) > 0 {
		rep = strings.Join(prefix, "/") + "/"
	}
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
		jst, err := utcToJst(lastModified)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if isDir {
			objects = append(objects, fmt.Sprintf("%s %10s  %s", jst, "-", key))
		} else {
			objects = append(objects, fmt.Sprintf("%s %10d  %s", jst, size, key))
		}
	}
	return
}

// Transform from UTC to JST
func utcToJst(utc time.Time) (string, error) {
	jst := utc.In(JST)
	return jst.Format("2006-01-2 15:03:04"), nil
}
