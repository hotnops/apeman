package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func populate_principal_blob(ctx context.Context, driver neo4j.DriverWithContext){
	fmt.Println("[*] Expanding principal blobs")
	query := "MATCH (a:AWSPrincipalBlob)"+
					" MATCH (b:AWSUser|AWSRole|AWSGroup|AWSIdentityProvider|AWSService)"+
					" WHERE b.arn =~ a.regex OR b.name =~ a.regex"+
					" MERGE (a) - [:ExpandsTo {layer: 2}] -> (b)"

	neo4j.ExecuteQuery(ctx, driver, query,
		nil, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))		

}


func populate_resource_blob(ctx context.Context, driver neo4j.DriverWithContext){
	fmt.Println("[*] Expanding resource blobs")
	query :=  "MATCH (a:AWSResourceBlob)" +
						" MATCH (b:UniqueArn)" +
						" WHERE b.arn =~ a.regex" +
						" MERGE (a) - [:ExpandsTo {layer: 2}] -> (b)"

	neo4j.ExecuteQuery(ctx, driver, query,
		nil, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))		

}


func populate_action_blob(ctx context.Context, driver neo4j.DriverWithContext){
	fmt.Println("[*] Expanding action blobs")
	query :=  "MATCH (a:AWSActionBlob)" +
						" MATCH (b:AWSAction)" +
						" WHERE b.name =~ a.regex" +
						" MERGE (a) - [:ExpandsTo {layer: 2}] -> (b)"

	neo4j.ExecuteQuery(ctx, driver, query,
		nil, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))		

}

func convert_variable_arn_to_regex(variable_arn string){
//TODO
}

func get_all_arn_nodes(ctx context.Context, driver neo4j.DriverWithContext){
//TODO
}

func populate_resource_types(ctx context.Context, driver neo4j.DriverWithContext){

	fmt.Println("[*] Expanding resource types")
	query :=  "MATCH (a:AWSResourceType)" +
						" MATCH (b:UniqueArn) WHERE (b.arn =~ a.regex)" +
						" MERGE (b)  - [:TypeOf {layer: 2}] -> (a)"
	
	neo4j.ExecuteQuery(ctx, driver, query,
		nil, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))		

}

func populate_arn_fields(ctx context.Context, driver neo4j.DriverWithContext){

	fmt.Println("[*] Populating ARN fields")

	query :=  "MATCH (u:UniqueArn)" +
						" WITH u, apoc.text.regexGroups(u.arn, 'arn:([^:]*):([^:]*):([^:]*):([^:]*):(.+)')[0] AS arn_parts" +
						" WHERE size(arn_parts) = 6" +
						" SET u.partition = arn_parts[1]," +
						" u.service = arn_parts[2]," +
            " u.region = arn_parts[3]," +
            " u.account_id = arn_parts[4],"+
            " u.resource = arn_parts[5]"

	neo4j.ExecuteQuery(ctx, driver, query,
		nil, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))	
}

func analyze_assume_roles(){

	fmt.Println("[*] Analyzing assume roles")
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")
	error_check(err)

	if resp.StatusCode == 200{
		fmt.Println("[*] Assume role analysis complete")

	}else{
		fmt.Println("[!] Could not analyze assume roles\n " + err.Error())

	}

}

func error_check(err error) {
	if err != nil {
		fmt.Println("Error with response:", err)
	}
}

func analyze(){

	 ctx := context.Background()
	//initialize database
	dbUri := "bolt://localhost:7687"
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.NoAuth())
	error_check(err)

	defer driver.Close(ctx)

	//analyze db
	populate_arn_fields(ctx,driver)
  populate_resource_types(ctx,driver)
  populate_action_blob(ctx,driver)
	populate_resource_blob(ctx,driver)
	populate_resource_blob(ctx,driver)

}

func main(){
	analyze()
}