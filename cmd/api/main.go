package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/irensaltali/object-storage-gateway/internal/api"
	"github.com/irensaltali/object-storage-gateway/internal/discovery"
	"github.com/irensaltali/object-storage-gateway/internal/storage"
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

	// Create object storage gateway
	gateway, err := storage.NewGateway(instances)
	if err != nil {
		log.Fatalf("failed to create gateway: %v", err)
	}
	defer gateway.Close()
	log.Printf("Gateway is ready to use")

	apiRouter := api.NewRouter(gateway)
	http.ListenAndServe(":8080", apiRouter)
	log.Printf("API server is running on port 8080")
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
