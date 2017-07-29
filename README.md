# ls3 : AWS S3 file explorer on CLI

![Working example](https://s3-ap-northeast-1.amazonaws.com/sugimoto/blog-images/ls3-example.gif)

## Installation

```
$ go get github.com/ysugimoto/ls3/...
```

It will installed `ls3` binary in your `$GOBIN`

## Usage

```
$ ls3 [options]
```

Show fiull usage for `ls3 -help`

```
$ ls3 -help

========================================================================
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
```

This tool can explore object file for drill-down and view (text file only) or download object.

## Author

Yoshiaki Sugimoto <sugimoto@wnotes.net>

## License

MIT License
