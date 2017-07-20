package main

import (
	"flag"
	"fmt"
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
		return
	}
	fmt.Printf("Retirving %s bucket...\n", selected)
	chooseObjects(selected, []string{})
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
		}
	} else if strings.HasSuffix(selected, "/") {
		selected = strings.TrimRight(selected[20:], "/")
		prefix = append(prefix, selected)
	} else {
		return
	}
	chooseObjects(bucket, prefix)
}

// Fromat object list from S3 response
func formatObjectList(result []*s3.Object, prefix []string) (objects []string) {
	if len(prefix) > 0 {
		objects = append(objects, "../")
	}
	unique := map[string]struct{}{}
	var rep string
	if len(prefix) > 0 {
		rep = strings.Join(prefix, "/") + "/"
	}
	for i := 0; i < len(result); i++ {
		key := *result[i].Key
		lastModified := *result[i].LastModified
		key = strings.Replace(key, rep, "", 1)
		if strings.Contains(key, "/") {
			spl := strings.Split(key, "/")
			if _, exists := unique[spl[0]]; exists {
				continue
			}
			unique[spl[0]] = struct{}{}
			key = spl[0] + "/"
		}
		jst, err := utcToJst(lastModified)
		if err != nil {
			fmt.Println(err)
			continue
		}
		objects = append(objects, fmt.Sprintf("%s %s", jst, key))
	}
	return
}

// Transform from UTC to JST
func utcToJst(utc time.Time) (string, error) {
	jst := utc.In(JST)
	return jst.Format("2006-01-2 15:03:04"), nil
}
