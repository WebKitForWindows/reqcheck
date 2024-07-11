# reqcheck
> Determine when a WebKit requirement has a new release

![example workflow](https://github.com/WebKitForWindows/reqcheck/actions/workflows/build.yml/badge.svg)

## Build

Build the binary with the following command:

```console
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

go build -v -a -tags netgo -o release\windows\amd64\reqcheck.exe ./cmd/reqcheck
```

## Docker

Build the Docker image with the following command:

```console
docker build --file docker\Dockerfile.windows.1809 --tag webkitdev/reqcheck .
```

Run the Docker image with the following command:

```console
docker run --rm `
    -v <path-to-requirements>:C:/WebKitRequirements `
    -w C:/WebKitRequirements `
    webkitdev/reqcheck vcpkg .
```
