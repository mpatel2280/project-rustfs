package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"rustfs-go-sdk/rustfs"
)

func main() {
	var (
		bucket      = flag.String("bucket", "", "RustFS bucket name")
		key         = flag.String("key", "", "Object key to read")
		out         = flag.String("out", "", "Optional file path to download the object into")
		list        = flag.Bool("list", false, "List objects in the bucket")
		listBuckets = flag.Bool("list-buckets", false, "List all buckets")
	)
	flag.Parse()

	ctx := context.Background()
	client, err := rustfs.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("create RustFS client: %v", err)
	}

	switch {
	case *listBuckets:
		buckets, err := client.ListBuckets(ctx)
		if err != nil {
			log.Fatalf("list buckets: %v", err)
		}
		for _, bucket := range buckets {
			fmt.Println(bucket.Name)
		}
	case *list:
		if *bucket == "" {
			log.Fatal("the -bucket flag is required when using -list")
		}
		objects, err := client.ListObjects(ctx, *bucket)
		if err != nil {
			log.Fatalf("list objects: %v", err)
		}
		for _, object := range objects {
			fmt.Printf("%s\t%d\n", object.Key, object.Size)
		}
	case *key != "" && *out != "":
		if *bucket == "" {
			log.Fatal("the -bucket flag is required when using -key")
		}
		if err := client.DownloadObject(ctx, *bucket, *key, *out); err != nil {
			log.Fatalf("download object: %v", err)
		}
		fmt.Fprintf(os.Stdout, "saved %s/%s to %s\n", *bucket, *key, *out)
	case *key != "":
		if *bucket == "" {
			log.Fatal("the -bucket flag is required when using -key")
		}
		data, err := client.ReadObject(ctx, *bucket, *key)
		if err != nil {
			log.Fatalf("read object: %v", err)
		}
		fmt.Fprint(os.Stdout, string(data))
	default:
		flag.Usage()
	}
}
