package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	storagev1 "github.com/logsquaredn/geocloud/api/storage/v1"
	"github.com/logsquaredn/geocloud/api/storage/v1/storagev1connect"
)

func main() {
	client := storagev1connect.NewStorageServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)

	stream := client.CreateStorage(context.Background())
	stream.RequestHeader().Add("X-Content-Type", "application/zip")
	stream.RequestHeader().Add("X-API-Key", "cus_LcKO8YPhzJZQgu")
	stream.RequestHeader().Add("X-Storage-Name", "phish3y_quake")

	f, err := os.Open("/home/phish3y/Documents/input/mi.zip")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, 8192)
	for {
		bRead, err := f.Read(buf)

		if err != nil {
			if err == io.EOF {
				err = stream.Send(&storagev1.CreateStorageRequest{
					Data: buf[:bRead],
				})
				if err != nil {
					panic(err)
				}

				break
			} else {
				panic(err)
			}
		}

		err = stream.Send(&storagev1.CreateStorageRequest{
			Data: buf[:bRead],
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sent %d bytes\n", bRead)
	}

	res, err := stream.CloseAndReceive()
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
