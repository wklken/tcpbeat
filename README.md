# Tcpbeat


Welcome to Tcpbeat.

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/wklken/tcpbeat`

## Reference

Since the tcp input plugin currently in the filebeat, not in the libbeat

base on filebeat tcp input plugin: https://github.com/elastic/beats/pull/62664

you can build on beat with tcp input, and make the event, push to the pipeline

## Getting Started with Tcpbeat

### Requirements

* [Golang](https://golang.org/dl/) 1.7

### Init Project
To get running with Tcpbeat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Tcpbeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/wklken/tcpbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Tcpbeat run the command below. This will generate a binary
in the same directory with the name tcpbeat.

```
make
```


### Run

To run Tcpbeat with debugging output enabled, run:

```
./tcpbeat -c tcpbeat.yml -e -d "*"
```


### Test

To test Tcpbeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `fields.yml` by running the following command.

```
make update
```


### Cleanup

To clean  Tcpbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Tcpbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/src/github.com/wklken/tcpbeat
git clone https://github.com/wklken/tcpbeat ${GOPATH}/src/github.com/wklken/tcpbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make package
```

This will fetch and create all images required for the build process. The hole process to finish can take several minutes.
