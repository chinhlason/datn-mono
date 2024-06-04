package elastic

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"os"
)

var ELASTIC_PASSWORD = os.Getenv("ELASTIC_PASSWORD")

func ConnectToElasticsearch() *elasticsearch.Client {
	cert, _ := os.ReadFile("/path/to/http_ca.crt")
	fmt.Println(cert)
	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Username: "elastic",
		Password: "NZBDUJQm6xYhufBzdMDm",
		CACert:   cert,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	return es
}
