# vreserve

Reserve space on a storage volume on *nix and POSIX systems.

## Why?

vreserve was originally part of a system on which many services were running.
Each service downloaded large files for data processing, then deleted the
files when processing completed.

Running out of disk space was a major problem in this system because the
services could not coordinate disk usage. The disk often filled up, and
all the services got stuck.

Enter core.

## What does it do?

vreserve keeps track of how much disk space is available. It lets other
services reserve and release space. Here's a sample use case:

* Ten different services each want to download and process large files.
  Let's say these files are 500GB each.
* Each asks vreserve if it can have a 500GB chunk of disk space.
* vreserve knows not only how much space is currently free on the 
  underlying disk, it also knows which processes have already asked
  for chunks of space for pending operations.
* vreserve can grant or deny each service's request for space based
  on this knowledge.
* If it grants a chunk of space, the requestor can go ahead and start
  their operation.
* If it denies a chunk of space, the requestor knows not to start now
  and to request again later, when more space may be available.
* This prevents `disk full` errors that block all services from 
  completing their work.
* When a service finishes its work and deletes the files it was using,
  it tells vreserve to release the disk space.
* vreserve can then grant that space to the next requestor.
* All grants are on a first-come, first-serve basis.

vreserve originally ran as a service, with requests coming in over a TCP
or HTTP/REST connection from other services. Since vreserve keeps all its
data in memory, it's important to have only one instance running.

The code in this repo includes only the vreserve core, not the service
layer. But it should be easy to wrap in Go's http server library.

## Building

`go build -o vreserve main.go`

## Server Usage

To run the server, simply run `go run main.go`.

If you want to specify a hostname, port and log file:

`go run main.go -H 0.0.0.0 -p 9999 -l /path/to/log`


## Client Usage

If external services are written in Go, you can use 
[VolumeClient](core/volume_client.go) to connect and request space.

From other languages, use the following REST calls, all of which return JSON.
Note that trailing slashes in URLs are requied.

**GET /ping/** Checks to see if the server is running. It should return:

```json
{
  "Succeeded":true,
  "ErrorMessage":"",
  "Data":null
}
```

**POST /reserve/** To reserve a chunk of disk space.

Requires POST params:

* path (string) - Path to the file for which you want to reserve space.
  vreserve finds the mountpoint of that path and determines if sufficient
  space is available on that device. Keep track of the path value, because
  you'll need to release it when you're done. Path should be absolute, so
  vreserve can figure out what device it's on.

* bytes (int) - The number of bytes you want to reserve.

Returns:

```json
{
  "Succeeded":true,
  "ErrorMessage":"",
  "Data":null
}
```

If vreserve thinks there's not enough space, it will set Succeeded to false
and supply an error message.

**POST /release/**

Requires POST params:

* path (string) - A path you previously reserved. 

If you previously reserved 100GB of space at this path, vreserve will 
update its internal ledger to indicate these 100GB are now free for 
other uses.

Returns:

```json
{
  "Succeeded":true,
  "ErrorMessage":"",
  "Data":null
}
```

**GET /report/?path=<path>**

Returns a report of all space reserved under the specified path.
You'll get different reports for different devices. For example, 
`path=/` may return something different than `path=/mnt/hdd1` if
the latter has a different mountpoint than the former.

Returns:

```json
{
    "Succeeded":true,
    "ErrorMessage":"",
    "Data":{
      "/data/abc":25165,
      "/data/xyz":998000
    }
}
```

The example above shows you've reserved 25,165 bytes of space at 
`/data/abc` and 998,000 bytes at `/data/xyz`, and neither has been
freed yet.

If you release one of the blocks by posting to the /release/ endpoint,
then call /report/ again, you'll see the released block has been
removed.

## Minimal curl test

Start a local server with `go run main.go`, then run the following:

```
curl http://localhost:8188/ping/

curl --data "path=/abc" --data "bytes=1000" http://localhost:8188/reserve/
curl --data "path=/def" --data "bytes=2000" http://localhost:8188/reserve/
curl --data "path=/xyz" --data "bytes=3000" http://localhost:8188/reserve/

curl http://localhost:8188/report/?path=/

curl --data "path=/abc" http://localhost:8188/release/
curl --data "path=/def" http://localhost:8188/release/

curl http://localhost:8188/report/?path=/
```

## Embedded/Library Usage

Let's assume service A is running on a machine with ten other services.
It wants to claim a 500GB chunk of space to download and process a file.
It does this:

```go

// Request 500 GB of space from vreserve at
// path "/path/to/file_1"
fiveHundredGB := (5 * 1024 * 1024 * 1024)
err = volume.Reserve("/path/to/file_1", fiveHundredGB)

if err == nil {
    // download file
    // do processing
    // delete file

    // Tell vreserve the space at that path is now free.
    // Once this is done, vreserve can offer it to others.
    volume.Release("/path/to/file_1")

} else {
    // Not enough space is available.
    // Requeue your task or tell requestor we can't do 
    // this now because there's not enough disk space.
}
```

Note that `/path/to/file` can be a path to any file or directory on any
mounted volume. vreserve figures out which volume that path is mounted
on and knows how much space is available on that volume.

This path also acts as a key. The caller should release the same path
it reserved.

General usage is simple. See the tests in [volume_test.go](volume_test.go).

## Caveats and Limitations

* vreserve uses an external call to `df` to determine volume mount points.
* df parsing may not be 100% accurate on all systems.
* Unknown mountpoints default to "/"
* vreserve does nothing to enforce volume reservations. If a process
  wants to go behind vreserve's back and eat up the whole disk, it
  can.
* vreserve works well enough when other processes, such as loggers, 
  are slowly filling up disk space in the background, but will fail
  if some process it knows nothing about starts filling up the whole
  disk.
* If you want lots of processes or services to coordinate disk usage
  through vreserve, they must **all** use vreserve to reserve and
  release disk space.

In short, think of vreserve like the conference room reservation system
in a shared office. It will manage reservations correctly, but it won't
keep bad actors out of the space you reserved.

In production, this has worked well enough for several years.

## Testing

`go test ./...`

