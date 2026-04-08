# RustFS Go SDK

This project provides a small Go SDK wrapper for reading files from RustFS buckets using the S3-compatible RustFS API.

It is designed for the flow:

- RustFS Browser
- bucket
- `bucketname`
- object/file read

RustFS documents its Go integration as standard S3-compatible access using the AWS SDK for Go v2.

## Features

- Create a RustFS client from config or environment variables
- List buckets
- List objects inside a bucket
- Read an object into memory
- Stream an object to a file

## Environment Variables

The CLI will automatically load values from a local `.env` file if present.

This project now includes:

```bash
RUSTFS_ENDPOINT_URL=http://localhost:9000
RUSTFS_REGION=us-east-1
RUSTFS_ACCESS_KEY_ID=rustfsadmin
RUSTFS_SECRET_ACCESS_KEY=rustfsadmin
```

You can still export them manually if needed:

```bash
export RUSTFS_ENDPOINT_URL="http://127.0.0.1:9000"
export RUSTFS_REGION="us-east-1"
export RUSTFS_ACCESS_KEY_ID="your-access-key"
export RUSTFS_SECRET_ACCESS_KEY="your-secret-key"
```

## Project Layout

- `rustfs/client.go`: SDK implementation
- `cmd/rustfs-reader/main.go`: example CLI for listing buckets/objects and reading files

## Run the Example CLI

List objects in a bucket:

```bash
go run ./cmd/rustfs-reader -bucket bucketname -list
```

Read a file from a bucket and print its contents:

```bash
go run ./cmd/rustfs-reader -bucket bucketname -key path/to/file.txt
```

Download a file from a bucket:

```bash
go run ./cmd/rustfs-reader -bucket bucketname -key path/to/file.txt -out ./file.txt
```

List buckets:

```bash
go run ./cmd/rustfs-reader -list-buckets
```

## Notes

- The SDK uses path-style S3 access by default, which matches RustFS documentation.
- If your RustFS deployment uses TLS or a custom region, update the environment variables accordingly.
