# bqwriter [![Go Workflow Status](https://github.com/OTA-Insight/bqwriter/workflows/Go/badge.svg)](https://github.com/OTA-Insight/bqwriter/actions/workflows/go.yml)&nbsp;[![GoDoc](https://godoc.org/github.com/OTA-Insight/bqwriter?status.svg)](https://godoc.org/github.com/OTA-Insight/bqwriter)&nbsp;[![Go Report Card](https://goreportcard.com/badge/github.com/OTA-Insight/bqwriter)](https://goreportcard.com/report/github.com/OTA-Insight/bqwriter)&nbsp;[![license](https://img.shields.io/github/license/OTA-Insight/bqwriter.svg)](https://github.com/OTA-Insight/bqwriter/blob/master/LICENSE.txt)&nbsp;[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/OTA-Insight/bqwriter?include_prereleases)](https://github.com/OTA-Insight/bqwriter/releases)

A Go package to write data into [Google BigQuery](https://cloud.google.com/bigquery/)
concurrently with a high throughput. By default [the InsertAll() API](https://cloud.google.com/bigquery/streaming-data-into-bigquery)
is used (REST API under the hood), but you can configure to use [the Storage Write API](https://cloud.google.com/bigquery/docs/write-api) (GRPC under the hood) as well.

The InsertAll API is easier to configure and can work pretty much out of the box without any configuration.
It is recommended to use the Storage API as it is faster and comes with a lower cost. The latter does however
require a bit more configuration on your side, including a Proto schema file as well.
See [the Storage example below](#Storage-Streamer) on how to do this.

A third API is available as well, and is a bit different than the other ones. A streamer using the batch API
expects the data to be written be an io.Reader of which encoded rows can be read from, in order to be batch
loaded into BigQuery. As such its purpose is different from the other 2 clients, and is meant more in an
environment where a lot of data has to be loaded at once into big query, rather than being part of a constant
high-speed high-throughput environment.

See https://cloud.google.com/bigquery/docs/batch-loading-data for more information about the Batch API,
and see [the Batch example below](#Batch-Streamer) on how to do create and use the Batch-driven Streamer.

Note that (gcloud) [Authorization](#Authorization) is implemented in the most basic manner.
[Write Error Handling](#write-error-handling) is currently not possible either, making the Streamer
a fire-and-forget BQ writer. Please read the sections on these to topics for more information
and please consult he [Contributing section](#Contributing) section explains how you can actively
help to get this supported if desired.

## Install

```go
import "github.com/OTA-Insight/bqwriter"
```

To install the packages on your system, do not clone the repo. Instead:

1. Change to your project directory:

```bash
cd /path/to/my/project
```

2. Get the package using the official Go tooling, which will also add it to your `Go.mod` file for you:

```bash
go get github.com/OTA-Insight/bqwriter
```

NOTE: This package is under development, and may occasionally make backwards-incompatible changes.

## Go Versions Supported

We currently support Go versions 1.15 and newer.

## Examples

In this section you'll find some quick examples to help you get started
together with the official documentation which you can find at <https://pkg.go.dev/github.com/OTA-Insight/bqwriter>.

The `Streamer` client is safe for concurrent use and can be used from as many go routines as you wish.
No external locking or other concurrency-safe mechanism is required from your side. To keep these examples
as small as possible however they are written in a linear synchronous fashion, but it is encouraged to use the
`Streamer` client from multiple go routines, in order to be able to write rows at a sufficiently high throughput.

Note that for the Batch-driven `Streamer` it is not abnormal to force it to run with a single worker routine.
The batch delay can also be disabled for it as no flushing is required for it anyhow.

Please also note that errors are not handled gracefully in these examples as ot keep them small and narrow in scope.

For extra reference you can also find some more examples, be it less pragmatic,
in the [./internal/test/integration](./internal/test/integration) directory.

### Basic InsertAll Streamer

```go
import (
    "context"

    "github.com/OTA-Insight/bqwriter"
)

// TODO: use more specific context
ctx := context.Background()

// create a BQ (stream) writer thread-safe client,
bqWriter, err := bqwriter.NewStreamer(
    ctx,
    "my-gcloud-project",
    "my-bq-dataset",
    "my-bq-table",
    nil, // use default config
)
if err != nil {
    // TODO: handle error gracefully
    panic(err)
}
// do not forget to close, to close all background resources opened
// when creating the BQ (stream) writer client
defer bqWriter.Close()

// You can now start writing data to your BQ table
bqWriter.Write(&myRow{Timestamp: time.UTC().Now(), Username: "test"})
// NOTE: only write one row at a time using `(*Streamer).Write`,
// multiple rows can be written using one `Write` call per row.
```

You build a `Streamer` client using optionally the `StreamerConfig` as you can see in the above example.
The entire config is optional and has sane defaults, but note that there is a lot you can configure in this config prior to actually building the streamer. Please consult the <https://pkg.go.dev/github.com/OTA-Insight/bqwriter#StreamerConfig> for more information.

The `myRow` structure used in this example is one way to pass in the information
of a single row to the `(*Streamer).Write` method. This structure implements the
[`ValueSaver`](https://pkg.go.dev/cloud.google.com/go/bigquery#ValueSaver) interface.
An example of this:

```go
import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
)

type myRow struct {
	Timestamp time.Time
	Username  string
}

func (mr *myRow) Save() (row map[string]bigquery.Value, insertID string, err error) {
	return map[string]bigquery.Value{
		"timestamp": civil.DateTimeOf(rr.Timestamp),
		"username":  mr.Username,
	}, "", nil
}
```

You can also pass in a `struct` directly and the schema will be inferred automatically
based on its public items. This flexibility has a runtime cost by having to apply reflection.

A raw `struct` can also be stored by using the [`StructSaver`](https://pkg.go.dev/cloud.google.com/go/bigquery#StructSaver) interface,
in which case you get the benefit of being able to write any kind of `struct` while at the same time being able to pass
in the to be used scheme already such that it doesn't have to be inferred and giving you exact controls for each field on top of that.

If you have the choice however than we do recommend to implement the `ValueSaver` for your row `struct` as this gives you the best of both worlds,
while at the same time also giving you the easy built-in ability to define a unique `insertID` per row which will help prevent potential duplicates
that can otherwise happen while retrying to write rows which have failed temporarily.

### Custom InsertAll Streamer

Using the same `myRow` structure from previous example,
here is how we create a `Streamer` client with a more
custom configuration:

```go
import (
    "context"

    "github.com/OTA-Insight/bqwriter"
)

// TODO: use more specific context
ctx := context.Background()

// create a BQ (stream) writer thread-safe client,
bqWriter, err := bqwriter.NewStreamer(
    ctx,
    "my-gcloud-project",
    "my-bq-dataset",
    "my-bq-table",
    &bqwriter.StreamerConfig{
        // use 5 background worker threads
        WorkerCount: 5,
        // ignore errors for invalid/unknown rows/values,
        // by default these errors make a write fail
        InsertAllClient: &bqwriter.InsertAllClientConfig{
             // Write rows fail for invalid/unknown rows/values errors,
             // rather than ignoring these errors and skipping the faulty rows/values.
             // These errors are logged using the configured logger,
             // and the faulty (batched) rows are dropped silently.
            FailOnInvalidRows:    true,
            FailForUnknownValues: true, 
        },
    },
)
if err != nil {
    // TODO: handle error gracefully
    panic(err)
}
// do not forget to close, to close all background resources opened
// when creating the BQ (stream) writer client
defer bqWriter.Close()

// You can now start writing data to your BQ table
bqWriter.Write(&myRow{Timestamp: time.UTC().Now(), Username: "test"})
// NOTE: only write one row at a time using `(*Streamer).Write`,
// multiple rows can be written using one `Write` call per row.
```

### Storage Streamer

If you can you should use the StorageStreamer. The InsertAll API is now considered legacy
and is more expensive and less efficient to use compared to the storage API.

Here follows an example on how you can create such a storage API driven BigQuery streamer.

```go
import (
    "context"

    "github.com/OTA-Insight/bqwriter"
    "google.golang.org/protobuf/reflect/protodesc"

    // TODO: define actual path to pre-compiled protobuf Go code
    "path/to/my/proto/package/protodata"
)

// TODO: use more specific context
ctx := context.Background()

// create proto descriptor to use for storage client
protoDescriptor := protodesc.ToDescriptorProto((&protodata.MyCustomProtoMessage{}).ProtoReflect().Descriptor())
// NOTE:
//  - storage writer API expects proto2 semantics, proto3 shouldn't be used (yet);
//  - the [normalizeDescriptor](https://pkg.go.dev/cloud.google.com/go/bigquery/storage/managedwriter/adapt#NormalizeDescriptor)
//    should be used to get a descriptor with nested types in order to have it work nicely with nested types;
//    - this means the line above would change to:
//      `protoDescriptor := adapt.NormalizeDescriptor((&protodata.MyCustomProtoMessage{}).ProtoReflect().Descriptor())`,
//      which does require the `"cloud.google.com/go/bigquery/storage/managedwriter/adapt"` package to be imported;
//  - known types cannot be used, you'll need to use type conversions instead
//    https://cloud.google.com/bigquery/docs/write-api#data_type_conversions,
//    e.g. int64 (micro epoch) instead of the known Google Timestamp proto type;

// create a BQ (stream) writer thread-safe client,
bqWriter, err := bqwriter.NewStreamer(
    ctx,
    "my-gcloud-project",
    "my-bq-dataset",
    "my-bq-table",
    &bqwriter.StreamerConfig{
        // use 5 background worker threads
        WorkerCount: 5,
        // create the streamer using a Protobuf message encoder for the data
        StorageClient: &bqwriter.StorageClientConfig{
            ProtobufDescriptor: protoDescriptor,
        },
    },
)
)
if err != nil {
    // TODO: handle error gracefully
    panic(err)
}
// do not forget to close, to close all background resources opened
// when creating the BQ (stream) writer client
defer bqWriter.Close()

// TOOD: populate fields of the proto message
msg := new(protodata.MyCustomProtoMessage)

// You can now start writing data to your BQ table
bqWriter.Write(msg)
// NOTE: only write one row at a time using `(*Streamer).Write`,
// multiple rows can be written using one `Write` call per row.
```

You must define the `StorageClientConfig`, as demonstrated in previous example,
in order to be create a Streamer client using the Storage API.
Note that you cannot create a blank `StorageClientConfig` or any kind of default,
as you are required to configure it with either a `bigquery.Schema` or a `descriptorpb.DescriptorProto`,
with the latter being preferred and used of the first.

The schema or Protobuf descriptor are used to be able to encode the data prior to writing
in the correct format as Protobuf encoded binary data.

- `BigQuerySchema` can be used in order to use a data encoder for the StorageClient
  based on a dynamically defined BigQuery schema in order to be able to encode any struct,
  JsonMarshaler, Json-encoded byte slice, Stringer (text proto) or string (also text proto)
  as a valid protobuf message based on the given BigQuery Schema;
- `ProtobufDescriptor` can be used in order to use a data encoder for the StorageClient
  based on a pre-compiled protobuf schema in order to be able to encode any proto Message
  adhering to this descriptor;

`ProtobufDescriptor` is preferred as you might have to pay a performance penalty
should you want to use the `BigQuerySchema` instead.

You can check out [./internal/test/integration/temporary_data_proto2.proto](./internal/test/integration/temporary_data_proto2.proto) for an example of a proto message that can be sent over the wire. The BigQuery
schema for that definition can be found in [./internal/test/integration/tmpdata.go](./internal/test/integration/tmpdata.go). Finally, you can get inspired by [./internal/test/integration/generate.go](./internal/test/integration/generate.go) to know how to generate the required Go code in order for you to configure your streamer with the right proto descriptor and being able to send rows of data using your proto definitions.

### Batch Streamer

The batch streamer can be used if you want to upload a big dataset of data to bigquery without any additional cost.

Here follows an example on how you can create such a batch API driven BigQuery client.

```go
import (
    "bytes"
    "context"
    "encoding/json"
    "path/filepath"

    "github.com/OTA-Insight/bqwriter"

    "cloud.google.com/go/bigquery"
)

func main() {
    ctx := context.Background()
	
    // By using new(bqwriter.BatchClientConfig) we will create a config with bigquery.JSON as default format
    // And the schema will be autodetected via the data.
    // Possible options are: 
    // - BigQuerySchema: Schema to use to upload to bigquery.
    // - SourceFormat: Format of the data we want to send.
    // - FailForUnknownValues: will treat records that have unknown values as invalid records.
    // - WriteDisposition: Defines what the write disposition should be to the bigquery table.
    batchConfig := new(bqwriter.BatchClientConfig)
	
    // create a BQ (stream) writer thread-safe client.
    bqWriter, err := bqwriter.NewStreamer(
        ctx,
        "my-gcloud-project",
        "my-bq-dataset",
        "my-bq-table",
        &bqwriter.StreamerConfig{
            BatchClient: batchConfig
        },
    )

    if err != nil {
        // TODO: handle error gracefully
        panic(err)
    }
		
    // do not forget to close, to close all background resources opened
    // when creating the BQ (stream) writer client
    defer bqWriter.Close()

    // a batch-driven BQ Streamer expects an io.Reader,
    // the source of the data isn't strictly defined as long as the source
    // format is supported. Usually you would fetch the data from large files
    // as this is where the batch client really shines
    files, err := filepath.Glob("/usr/joe/my/data/path/exported_data_*.json")
    if err != nil {
        // TODO: handle error gracefully
        panic(err)
    }
    for _, fp := range files {
        file, err := os.Open(fp)
        if err != nil {
            // TODO: handle error gracefully
            panic(err)
        }

        // Write the data to bigquery.
        err := bqWriter.Write(file)
        if err != nil {
            // TODO: handle error gracefully
            panic(err)
        }
    }
}
```

You must define the `BatchClientConfig`, as demonstrated in previous example,
in order to create a Batch client.

Note that you cannot create a blank `BatchClientConfig` or any kind of default,
as you are required to configure it with at least a `SourceFormat`.

When using the Json format make sure the casing of your fields matches exactly the
fields defined in your BigQuery schema of the desired target table. While field names
are normally considered case insensitive, they do seem to cause "duplicate field" issues
as part of the batch load io.Reader decode process such as the following ones:

```
Job returned an error status {Location: "query"; Message: "Duplicate(Case Insensitive) field names: value and Value. Table: tmp_2e6895b9_b44b_4b5c_9941_def9a10e85d5_source"; Reason: "invalidQuery"}
```

Fix the casing of your json definition and this error should go away.

**BatchClientConfig options**:

- `BigQuerySchema` can be used in order to use a data encoder for the batchClient
  based on a dynamically defined BigQuery schema in order to be able to encode any struct,
  JsonMarshaler, Json-encoded byte slice, Stringer (text proto) or string (also text proto)
  as a valid protobuf message based on the given BigQuery Schema.
  
  The `BigQuerySchema` is required for all `SourceFormat` except for `bigquery.CSV` and `bigquery.JSON` as these
  2 formats will auto detect the schema via the content.

- `SourceFormat` is used to define the format that the data is that we will send.
  Possible options are:
  - `bigquery.CSV`
  - `bigquery.Avro`
  - `bigquery.JSON`
  - `bigquery.Parquet`
  - `bigquery.ORC`

- `FailForUnknownValues` causes records containing such values
  to be treated as invalid records.
  
  Defaults to false, making it ignore any invalid values, silently ignoring these errors,
  and publishing the rows with the unknown values removed from them.

- `WriteDisposition` can be used to define what the write disposition should be to the bigquery table.
  Possible options are:
    - `bigquery.WriteAppend`
    - `bigquery.WriteTruncate`
    - `bigquery.WriteEmpty`
  
  Defaults to `bigquery.WriteAppend`, which will append the data to the table.

#### Future improvements

Currently, the package does not support any additional options that the different `SourceFormat` could have, feel free to
open a feature request to add support for these.

## Authorization

The streamer client will use [Google Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials) for authorization credentials used in calling the API endpoints.
This will allow your application to run in many environments without requiring explicit configuration.

Please open an issue should you require more advanced forms of authorization. The issue should come with an example,
a clear statement of intention and motivation on why this is a useful contribution to this package. Even if you wish
to contribute to this project by implementing this patch yourself, it is none the less best to create an issue prior to it,
such that we can all be aligned on the specifics. Good communication is key here.

It was a choice to not support these advanced authorization methods for now. The reasons being that the package
authors didn't have a need for it and it allowed to keep the API as simple and small as possible. There however some
advanced authorizations still possible:

- Authorize using [a custom Json key file path](https://cloud.google.com/iam/docs/creating-managing-service-account-keys);
- Authorize with more control by using the [`https://pkg.go.dev/golang.org/x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2) package
  to create an `oauth2.TokenSource`;

To conclude. We currently do not support advanced ways for Authorization, but we're open to include support for these,
if there is sufficient interest for it. The [Contributing section](#Contributing) section explains how you can actively
help to get this supported if desired.

## Instrumentation

We currently support the ability to implement your logger which can be used instead of the standard logger which prints
to STDERR. It is used for debug statements as well as unhandled errors. Debug statements aren't used everywhere, any unhandled error that isn't propagated is logged using the used logger.

> You can find the interface you would need to implement to support your own Logger at
> <https://godoc.org/github.com/OTA-Insight/bqwriter/log#Logger>.

The internal client of the Storage-API driven Streamer also provides the tracking of stats regarding its GRPC functionality.
This is implemented and utilized via the <https://github.com/census-instrumentation/opencensus-go> package.

If you use `OpenCensus` for your own project it will work out of the box.

In case your project uses another data ingestion system you can none the less get these statistics within your system
of choice by registering an exporter which exports the stats to the system used by your project. Please see
https://github.com/census-instrumentation/opencensus-go#views as a starting point on how to register a view yourself.
OpenCensus comes with a bunch of exporters already, all listed in https://github.com/census-instrumentation/opencensus-go#exporters.
You can however also implement your own one.

The official google cloud API will most likely switch to OpenCensus's successor OpenTelemetry once the latter becomes stable.
For now however it is OpenCensus that is used.

Note that this extra form of instrumentation is only applicable to a Streamer using the Storage API. The InsertAll-
and Batch-driven Streamers do not provide any form of stats tracking.

Please see also <https://github.com/googleapis/google-cloud-go/issues/5100#issuecomment-966461501> for more information
on how you can hook up a built-in or your own system into the tracking system for any storage API driven streamer.

## Write Error handling

The current version of the bqwriter is written with a fire-and-forget philosophy in mind.
Actual write errors occur on async worker goroutines and are only logged. Already today,
you can plugin your own logger implementation in order to get these logs in your alerting systems.

Please file a detailed feature request with a real use case as part of the verbose description
should you be in need of being able to handle errors.

One possible approach would be to allow a channel or callback to be defined in the `StreamerConfig`
which would get a specific data structure for any write failure. This could contain the data which failed to write,
any kind of offset/insertID as well as the actual error which occurred. The details would however to be worked out
as part of the proposal.

Besides a valid use case to motivate this proposal we would also need to think carefully about how
we can make the returned errors actionable. Returning it only to allow the user to log/print it is a bit silly,
as that is anyway already the behavior today. The real value from this proposal would come from the fact that
the data can be retried to be inserted (if it makes sense within its context, as defined by at the very least the error type),
and done so in an easy and safe manner, and with actual aid to help prevent duplicates. The Google Cloud API provides for this purpose
the offsets and insertID's, but the question is how we would integrate this and also to double check that this really does prevent
duplicates or not.

The [Contributing section](#Contributing) section explains how you can actively
help to get this supported if desired.

## Contributing

Contributions are welcome. Please, see the [CONTRIBUTING](/CONTRIBUTING.md) document for details.

Please note that this project is released with a Contributor Code of Conduct.
By participating in this project you agree to abide by its terms.
See [Contributor Code of Conduct](/CONTRIBUTING.md#contributor-code-of-conduct) for more information.

## Developer Instructions

As a developer you need to agree to the
[Contributor Code of Conduct](/CONTRIBUTING.md#contributor-code-of-conduct) for more information.
See [the previous Contributing section](#contributing) for more info in regards of contributing to this project.
In this section we'll also assume that you've read & understood the [Install](#install) and [Examples](#examples) sections.

Please take your time and complete the forms with sufficient details when filing issues and proposals. Pull requests (PRs) should only be created once a related issue/proposal has been created and agreed upon. Also take your time and complete the PR description with sufficient detail when you're ready to create a PR.

### Tests

Using [GitHub actions](./.github/workflows/go.yml) this codebase is being tested automatically for each commit/PR.
- `$ go test -v ./...`:
  - run against the Min and Max Go versions
  - all tests are expected to pass
- `$ golangci-lint run`:
  - run against latest Go version only
  - is expected to generate no warnings or errors of any kind

For each contribution that you do you'll have to make sure that all these tests pass.
Please do not modify any existing tests unless required because some kind of breaking change. If you do have to modify (or delete) existing tests than please document this in full detail with proper motivation as part of your PR description. Ensure your added and modified code is also sufficiently tested and covered.

Next to this, the maintainers of this repository (see [CODEOWNERS](CODEOWNERS)) also run [integration tests](./internal/test/integration) against a real production-like BigQuery table within the actual Google Cloud infrastructure. These test the streamer for all implementations: `insertAll`, `storage`, `storage-json` (a regular `storage` client but using a bigQuery.Schema as to be able to insert JsonMarshalled data) and `batch`.

You can run these tests yourself as well using the following internal cmd tool:

```batch
$ go run ./internal/test/integration --help
Usage of ./internal/test/integration/tmp/exe:
  -dataset string
        BigQuery dataset to write data to (default "benchmarks_bqwriter")
  -debug
        enable to show debug logs
  -iterations int
        how many values to write to each of the different streamer tests (default 100)
  -project string
        BigQuery project to write data to (default "oi-bigquery")
  -streamers string
        csv of streamers to test, one or multiple of following options: insertall, storage, storage-json, batch
  -table string
        BigQuery table to write data to (default "tmp")
  -workers int
        how many workers to use to run tests in parallel (default 12)
```

Most likely you'll need to pass the `--project`, `--dataset` and `--table` flag to use 
a BigQuery table for which you have sufficient permissions and that is used
only for temporary testing purposes such as these.

Running these tests yourself is not required as part of a contribution,
but it can be run by you in case you are interested in doing so for whatever reason.
