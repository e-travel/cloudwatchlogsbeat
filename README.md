[![Travis Build Status](https://travis-ci.org/e-travel/cloudwatchlogsbeat.svg?branch=master)](https://travis-ci.org/e-travel/cloudwatchlogsbeat)
[![Go Report
Card](https://goreportcard.com/badge/github.com/e-travel/cloudwatchlogsbeat)](https://goreportcard.com/report/github.com/e-travel/cloudwatchlogsbeat)

# Cloudwatchlogsbeat

Cloudwatchlogsbeat is a [beat](https://www.elastic.co/products/beats) for
the [elastic stack](https://www.elastic.co/products). Its purpose is
to harvest data from AWS Cloudwatch Log Groups and ship them to a
variety of sinks that include logstash, elasticsearch etc.

**Disclaimer**: the beat is production-tested and is currently being
used to harvest thousands of stream events per minute. However, please
keep in mind that there may be bugs and that we accept no
responsibility for any kind of damage that may occur as a result.

# Description

Cloudwatchlogsbeat operates by monitoring a set of AWS Cloudwatch Log
Groups specified in
its [configuration](cloudwatchlogsbeat.full.yml). The log groups are
periodically probed for new streams which are then polled for new
events.

The beat if fully concurrent in terms of the monitored log groups and
streams and makes use of AWS SDK's exponential back-off retry policy
for all its requests to the AWS APIs to avoid throttling errors. Any
throttling errors that will inevitably occur (due to maximum retries
for example) are dealt with gracefully without losing stream events
(stream monitoring is resumed from where it left off).

The state of the beat is saved in a user-specified S3 bucket on a
per-stream basis. This way, the beat knows what is the last event that
was harvested per stream and can resume its operation once restarted.

# Setup / Installation

First of all, make sure that you have
a [working go installation](https://golang.org/doc/install) (this
includes a valid `$GOPATH`).

The following steps are necessary for a working installation:

    $ go get -u github.com/e-travel/cloudwatchlogsbeat
    $ cd $GOPATH/src/github.com/e-travel/cloudwatchlogsbeat
    $ go build -i # builds the beat and builds/installs the dependencies
    $ ./cloudwatchlogsbeat -e -d '*'

# AWS configuration

Cloudwatchlogsbeat authenticates with AWS services using
the
[standard AWS guidelines](https://aws.amazon.com/blogs/security/a-new-and-standardized-way-to-manage-credentials-in-the-aws-sdks/). This
means that the following environmental variables need to be set for
the program to use:

    AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY

Alternatively, if there are profiles setup in the file
`~/.aws/credentials`, the following environmental variables are
necessary:

    AWS_PROFILE

The AWS region must be set in the beat's configuration file.

If the beat is deployed to an EC2 instance, there's also the option of
an IAM Role that is attached to the EC2 instance. In this case, the
actions that must be allowed in the IAM policy document are as
follows:

```
logs:DescribeLogGroups
logs:DescribeLogStreams
logs:GetLogEvents
logs:FilterLogEvents
logs:Describe*
```
plus `s3:*` on the S3 bucket resource.

# Tests

The beat's tests can be executed as follows:

    $ cd beater
    $ go test -v

# Contributing

Bug reports and pull requests are welcome on GitHub at
https://github.com/e-travel/cloudwatchlogsbeat. This project is
intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to
the [Contributor Covenant](http://contributor-covenant.org) code of
conduct.


# License

The beat is available as open source under the terms of
the [MIT License](http://opensource.org/licenses/MIT).
