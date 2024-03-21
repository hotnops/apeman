package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hotnops/apeman/analyze"
	"github.com/hotnops/apeman/src/config"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

var DriverName = "neo4j"

func main() {

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)

	configFilePath := "exampleConfig.json"
	ctx := context.Background()

	neo4j.Init()

	bhCfg, err := config.GetConfiguration(configFilePath)
	if err != nil {
		log.Fatalf("Unable to read configuration %s: %v", configFilePath, err)
	}

	dawgsCfg := dawgs.Config{
		DriverCfg:            bhCfg.Neo4J.Neo4jConnectionString(),
		TraversalMemoryLimit: size.Size(bhCfg.TraversalMemoryLimit) * size.Gibibyte,
	}

	graphDatabase, err := dawgs.Open("neo4j", dawgsCfg)
	if err != nil {
		log.Fatalf("Failed to open graph database")
	}

	roleToPDMap, err := analyze.GetSelfContainedTierZeroRoles(ctx, graphDatabase)
	if err != nil {
		log.Fatalf("Error")
	}

	log.Printf("Found %d statements", len(roleToPDMap))

	graphDatabase.ReadTransaction(ctx, func(tx graph.Transaction) error {
		roles, err := analyze.GetAWSRoles(tx)
		if err != nil {
			log.Fatalf("GetAWSRoles:  %s\n", err)
		}
		roles.Each(func(item uint32) (bool, error) {
			log.Printf("Role ID: %d\n", item)
			return true, nil
		})
		return err
	})
}
