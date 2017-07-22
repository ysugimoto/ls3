package main

import (
	"flag"
	"fmt"
	"os"

	"os/signal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

// Main function
func main() {
	sess := session.Must(session.NewSession())
	var conf *aws.Config
	if cli.env {
		conf = configFromEnv(cli.region)
	} else {
		conf = configFromProfile(cli.profile, cli.region)
	}

	service := s3.New(sess, conf)
	app, err := NewApp(service, cli.bucket)
	if err != nil {
		fmt.Println(err)
		return
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		for {
			select {
			case <-sc:
				app.Terminate()
				os.Exit(1)
			}
		}
	}()

	if err := app.Run(); err != nil {
		app.Terminate()
		os.Exit(1)
	}
	app.Terminate()
}
