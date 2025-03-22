package main

import (
	"fmt"

	"github.com/c4me-caro/drive/cmd/api"
	"github.com/c4me-caro/drive/database"
)

func main() {
	client, err := database.ConnectDB("mongodb://localhost:27017")
	if err != nil {
		fmt.Println(err)
		return
	}

	worker := database.NewDriveWorker(client, "testing", "driveapi")
	err = worker.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	server := api.NewApiServer(":8080", worker)
	if err := server.Run(); err != nil {
		fmt.Println(err)
		return
	}
}
