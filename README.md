# Cloudwatchlogsbeat

Cloudwatchlogsbeat is a [beat](https://www.elastic.co/products/beats) for
the [elastic stack](https://www.elastic.co/products). Its purpose is
to harvest data from AWS Cloudwatch Log Groups and ship them to
logstash/elasticsearch.

# Setup

First of all, make sure that you have
a [working go installation](https://golang.org/doc/install).

Dependency management is done using [glide](https://glide.sh/), so
make sure that it is installed.

The following steps are necessary for a working installation:

    $ glide install
    $ go build -i # builds and installs the dependencies
    $ go build -v # builds the beat
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

The AWS region can be set in the beat's configuration file.

# Tests

The beat's tests can be executed as follows:

    $ cd beater
    $ go test -v

# Deployment

## Build

Consider building the project using

    $ go build -ldflags="-w -s"

The generated executable is about 50% smaller.

## Elasticsearch

The `elasticsearch` host is `localhost` by default but can be
overriden from the command line as follows:

    $ ./cloudwatchlogsbeat -E output.elasticsearch.hosts=http://elastisearch.somewhere.org:9200

# Further Work

## Multiple Registries

S3 registry has been implemented

Pending:
* in-memory (default for local harvesting)

## Concurrency

streams and groups need to communicate through a channel on the
following occasions:

* stream has expired (last event is too old - group needs to cleanup
  the stream)
* stream has a fatal error (group needs to cleanup
  the stream)
* SIGTERM interrupt has been received - stream needs to cleanup / save
  state

the same is true for group and manager communication (but less of a
priority)

### Tests

Write tests using aws mocking libraries
for
[cloudwatchlogs](https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatchlogs/cloudwatchlogsiface/) and
[dynamodb](https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/dynamodbiface/)
