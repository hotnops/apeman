package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// PolicyEditorConfig represents the top-level JSON structure
type PolicyEditorConfig struct {
	PolicyEditorConfig_1 []string                  `json: "app.PolicyEditorConfig"`
	ConditionOperators   []string                  `json:"conditionOperators"`
	ConditionKeys        []string                  `json:"conditionKeys"`
	ServiceMap           map[string]ServiceDetails `json:"serviceMap"`
}

// ServiceDetails represents the details of each service in the service map
type ServiceDetails struct {
	//StringPrefix string   `json:"StringPrefix"`
	Actions     []string `json:"Actions"`
	ARNFormat   string   `json:"ARNFormat"`
	ARNRegex    string   `json:"ARNRegex"`
	HasResource bool     `json:"HasResource"`
}

type Content struct {
	Title    string    `json:"title"`
	Href     string    `json:"href"`
	Contents []Content `json:"contents,omitempty"`
}

type Root struct {
	Contents []Content `json:"contents,omitempty"`
}

type Actions struct {
	Action        string
	Description   string
	Accesslevel   string
	ResourceType  string
	ConditionKeys string
}

func printSubsections(content Content) {
	for _, subContent := range content.Contents {
		fmt.Printf("Title: %s\n", subContent.Title)
		fmt.Printf("Href: %s\n", subContent.Href)
	}
}

func get_service_metadata(url string) (*PolicyEditorConfig, error) {

	response, err := http.Get(url)
	error_check(err)

	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	error_check(err)

	// Define the structure to store the JSON data
	var config PolicyEditorConfig

	// Unmarshal the JSON data into the struct
	output := regexp.MustCompile("app.PolicyEditorConfig=")
	edit := output.ReplaceAllString(string(body), "")

	err = json.Unmarshal([]byte(edit), &config)
	if err != nil {
		error_check(err)
	}

	return &config, nil
}

func get_services_json_href(url string) map[string]string {

	//web request
	response, err := http.Get(url)
	error_check(err)

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	error_check(err)

	//parse json data
	var contents Root
	err = json.Unmarshal(body, &contents)
	if err != nil {
		fmt.Println("Error with unmarshal:", err)
	}

	//var serviceList []string

	serviceList := make(map[string]string)

	for _, data := range contents.Contents {
		//fmt.Println(data.Contents)
		for _, y := range data.Contents {
			//printSubsections(y)
			for _, subContent := range y.Contents {

				serviceList[subContent.Title] = subContent.Href
			}

		}
	}

	return serviceList

}

func get_service_dict(link string) ([]Actions, error) {
	//grab href from services json file, and pull the actions from the link

	response, err := http.Get(link)
	error_check(err)

	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	error_check(err)

	var details []Actions
	re := regexp.MustCompile(" +")

	//table #may change with each html response so I should figure out how to add programatically
	doc.Find("table").Eq(0).Each(func(column int, tr *goquery.Selection) {
		var detail Actions
		tr.Find("tr").Each(func(col int, td *goquery.Selection) {
			table_length := td.Find("td, th").Length()
			td.Find("td").Each(func(col int, td *goquery.Selection) {

				if col == 0 && td.Text() != "" {
					if table_length == 3 {

						detail.ResourceType = strings.TrimSpace(td.Text())
					} else if table_length == 4 {
						detail.Description = strings.TrimSpace(td.Text())
					} else {

						detail.Action = strings.TrimSpace(td.Text())
					}

				} else if col == 1 && td.Text() != "" {
					if table_length == 3 {

						text := strings.TrimSpace(td.Text())
						text = re.ReplaceAllString(text, " ")
						text = strings.ReplaceAll(text, "\n", " ")
						detail.ConditionKeys = text
					} else {

						detail.Description = strings.TrimSpace(td.Text())
					}
				} else if col == 2 && td.Text() != "" {

					detail.Accesslevel = strings.TrimSpace(td.Text())

				} else if col == 3 {

					detail.ResourceType = strings.TrimSpace(td.Text())

				} else if col == 4 {

					text := strings.TrimSpace(td.Text())
					text = strings.ReplaceAll(text, "\n", ", ")
					detail.ConditionKeys = text

				}
			})

			if detail.Action != "" || detail.Description != "" || detail.ResourceType != "" || detail.ConditionKeys != "" || detail.Accesslevel != "" {
				details = append(details, detail)
			}
		})
	})
	return details, nil
}

func aws_initialize(ctx context.Context, output_directory string) {

	dbUri :=  "bolt://localhost:7687"


	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.NoAuth())

	error_check(err)

defer driver.Close(ctx)

}

func write_data_to_csv(filename string, data []string) {
	//output this all to a csv

	file, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	//Title of column for CSV
	header := []string{"name"}
	writer.Write(header)

	for _, value := range data {
		err := writer.Write([]string{value})
		if err != nil {
			panic(err)
		}
	}

}

func error_check(err error) {
	if err != nil {
		fmt.Println("Error with response:", err)
	}
}

func main() {
	//urls that will be used to grab services, policies, and actions
	service_url := "https://awspolicygen.s3.amazonaws.com/js/policies.js"
	base_url := "https://docs.aws.amazon.com/service-authorization/latest/reference/"
	service_json_url := base_url + "toc-contents.json"
	
	//context for neo4j database
	ctx := context.Background()
	
	output_directory, err := os.Getwd()

	services_metadata, err := get_service_metadata(service_url)
	error_check(err)

	actionMap := make(map[string]Actions)
	serviceMap := make(map[string]interface{})

	//awsglobalconditionkeys.csv
	write_data_to_csv("awsglobalconditionkeys.csv", services_metadata.ConditionKeys)
	fmt.Println("[*] Condition keys written to awsglobalconditionkeys.csv")

	//awsoperators.csv
	write_data_to_csv("awsoperators.csv", services_metadata.ConditionOperators)
	fmt.Println("[*] Operators written to awsoperators.csv")

	fmt.Println("[*] Gathering URLs of services")
	href := get_services_json_href(service_json_url)

	fmt.Println("[*] Gathering Metadata of services")
	for title, url := range href {
		action_metadata, err := get_service_dict(base_url + url)
		error_check(err)
		actionMap = make(map[string]Actions)	

		for _, items := range action_metadata {
			actionMap[items.Action] = items
		}
		serviceMap[title] = actionMap
		fmt.Println(serviceMap[title])
	}

	//write json file
	fmt.Println("[*] Writing to json file")

	//convert servicemap to json
	json_output, err := json.MarshalIndent(serviceMap, "", "  ")
	error_check(err)

	//create and write to file
	aws_scheme_filewriter, err := os.Create("awsschema-test.json")
	error_check(err)

	defer aws_scheme_filewriter.Close()
	aws_scheme_filewriter.Write(json_output)

	fmt.Println("[*] Data written to awsschema.json")

	//initialize database
	aws_initialize(ctx,output_directory)

}
