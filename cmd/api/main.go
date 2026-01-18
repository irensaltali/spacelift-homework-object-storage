package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/spacelift-io/object-storage-gateway/internal/api"
	"github.com/spacelift-io/object-storage-gateway/internal/discovery"
)

func main() {
	printCredits()

	// Discovery Minio instances
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	instances, err := discovery.DiscoverInstances(ctx)
	cancel()

	if err != nil {
		println("Error discovering Minio instances:", err.Error())
		return
	}

	if len(instances) == 0 {
		println("No Minio instances found.")
		return
	}

	log.Printf("discovered %d minio instance(s)", len(instances))
	for _, inst := range instances {
		log.Printf("  - instance %s at %s:%s", inst.ID, inst.Host, inst.Port)
	}

	apiRouter := api.NewRouter()
	http.ListenAndServe(":8080", apiRouter)
}

func printCredits() {
	println(`
   /$$
   /_/
 /$$$$$$                                      /$$$$$$            /$$   /$$               /$$|
|_  $$_/                                     /$$__  $$          | $$  | $$              | $$|
  | $$    /$$$$$$   /$$$$$$  /$$$$$$$       | $$  \__/  /$$$$$$ | $$ /$$$$$$    /$$$$$$ | $$ /$$
  | $$   /$$__  $$ /$$__  $$| $$__  $$      |  $$$$$$  |____  $$| $$|_  $$_/   |____  $$| $$| $$
  | $$  | $$  \__/| $$$$$$$$| $$  \ $$       \____  $$  /$$$$$$$| $$  | $$      /$$$$$$$| $$| $$
  | $$  | $$      | $$_____/| $$  | $$       /$$  \ $$ /$$__  $$| $$  | $$ /$$ /$$__  $$| $$| $$
 /$$$$$$| $$      |  $$$$$$$| $$  | $$      |  $$$$$$/|  $$$$$$$| $$  |  $$$$/|  $$$$$$$| $$| $$
|______/|__/       \_______/|__/  |__/       \______/  \_______/|__/   \___/   \_______/|__/|__/


                                                   /$$ /$$  /$$$$$$   /$$           /$$                                                                           /$$      
                                                  | $$|__/ /$$__  $$ | $$          | $$                                                                          | $$      
  /$$$$$$$  /$$$$$$   /$$$$$$   /$$$$$$$  /$$$$$$ | $$ /$$| $$  \__//$$$$$$        | $$$$$$$   /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$  /$$  /$$  /$$$$$$   /$$$$$$ | $$   /$$
 /$$_____/ /$$__  $$ |____  $$ /$$_____/ /$$__  $$| $$| $$| $$$$   |_  $$_/        | $$__  $$ /$$__  $$| $$_  $$_  $$ /$$__  $$| $$ | $$ | $$ /$$__  $$ /$$__  $$| $$  /$$/
|  $$$$$$ | $$  \ $$  /$$$$$$$| $$      | $$$$$$$$| $$| $$| $$_/     | $$          | $$  \ $$| $$  \ $$| $$ \ $$ \ $$| $$$$$$$$| $$ | $$ | $$| $$  \ $$| $$  \__/| $$$$$$/ 
 \____  $$| $$  | $$ /$$__  $$| $$      | $$_____/| $$| $$| $$       | $$ /$$      | $$  | $$| $$  | $$| $$ | $$ | $$| $$_____/| $$ | $$ | $$| $$  | $$| $$      | $$_  $$ 
 /$$$$$$$/| $$$$$$$/|  $$$$$$$|  $$$$$$$|  $$$$$$$| $$| $$| $$       |  $$$$/      | $$  | $$|  $$$$$$/| $$ | $$ | $$|  $$$$$$$|  $$$$$/$$$$/|  $$$$$$/| $$      | $$ \  $$
|_______/ | $$____/  \_______/ \_______/ \_______/|__/|__/|__/        \___/        |__/  |__/ \______/ |__/ |__/ |__/ \_______/ \_____/\___/  \______/ |__/      |__/  \__/
          | $$                                                                                                                                                             
          | $$                                                                                                                                                             
          |__/                                                                                                                                                             
`)
}
