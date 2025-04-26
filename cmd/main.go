package main

import (
	"fmt"
	"os"

	"github.com/c4me-caro/drive/cmd/api"
	"github.com/c4me-caro/drive/database"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	
	client, err := database.ConnectDB(os.Getenv("MONGO_URI"))
	if err != nil {
		fmt.Println(err)
		return
	}

	worker := database.NewDriveWorker(client, os.Getenv("MONGO_DB"))
	err = worker.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
  
  fmt.Printf("Server running on: %s", os.Getenv("ADDRESS"))
	server := api.NewApiServer(os.Getenv("ADDRESS"), worker)
	if err := server.Run(); err != nil {
		fmt.Println(err)
		return
	}
}
