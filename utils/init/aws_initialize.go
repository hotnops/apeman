package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	StringPrefix string   `json:"StringPrefix"`
	Actions      []string `json:"Actions"`
	ARNFormat    string   `json:"ARNFormat"`
	ARNRegex     string   `json:"ARNRegex"`
	HasResource  bool     `json:"HasResource"`
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
	Action      string
	Description string
	Accesslevel string
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
	error_check(err)

	// Print the parsed data

	/*
		fmt.Println("Condition Operators:", config.ConditionOperators)
		fmt.Println("Condition Keys:", config.ConditionKeys)


			for service, details := range config.ServiceMap {

				fmt.Println("Service:", service)
				fmt.Println("  String Prefix:", details.StringPrefix)
				fmt.Println("  Actions:", details.Actions)
				fmt.Println("  ARN Format:", details.ARNFormat)
				fmt.Println("  ARN Regex:", details.ARNRegex)
				fmt.Println("  Has Resource:", details.HasResource)
			}
	*/
	return &config, nil
}

func get_services_json_href(url string) []string {

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

	var hrefList []string

	for _, data := range contents.Contents {
		//fmt.Println(data.Contents)
		for _, y := range data.Contents {
			//printSubsections(y)
			for _, subContent := range y.Contents {
				hrefList = append(hrefList, subContent.Href)
				//fmt.Println(subContent.Href)
			}

		}
	}

	return hrefList

}

func html_table_to_json(html_response string) {
	//when pulling aws asctions table, convert the table to json
	//web request

}

func get_service_dict(link string) {
	//grab href from services json file, and pull the actions from the link

	response, err := http.Get(link)
	error_check(err)

	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	error_check(err)

	//table #may change with each html response so I should figure out how to add programatically
	doc.Find("table").Eq(0).Each(func(column int, tr *goquery.Selection) {
		tr.Find("tr").Each(func(col int, td *goquery.Selection) {
			table_length := td.Find("td, th").Length()
			td.Find("td").Each(func(col int, td *goquery.Selection) {

				if col == 0 && td.Text() != ""{
					if table_length == 3 {
						fmt.Printf("Resource Type:  %s\n", strings.TrimSpace(td.Text()))
					} else {
						fmt.Printf("Actions:  %s\n", strings.TrimSpace(td.Text()))
					}

				} else if col == 1 && td.Text() != "" {
					if table_length == 3 {
						fmt.Printf("Condition Keys:  %s\n", strings.TrimSpace(td.Text()))
					} else {
						fmt.Printf("Description:  %s\n", strings.TrimSpace(td.Text()))
					}

				} else if col == 2 && td.Text() != ""{
					fmt.Printf("Access Level:  %s\n", strings.TrimSpace(td.Text()))

				} else if col == 3 {
					fmt.Printf("Resource Type:  %s\n", strings.TrimSpace(td.Text()))

				} else if col == 4 {
					fmt.Printf("Condition Keys:  %s\n", strings.TrimSpace(td.Text()))

				} else {
					fmt.Printf("\n")
				}
			})

		})

	})

}

func aws_initialize() {
	// save for later
}

func write_data_to_csv() {
	//output this all to a csv

}

func error_check(err error) {
	if err != nil {
		fmt.Println("Error with response:", err)
	}
}

func json_error_check(err error) {

}

func response_error_check(err error) {

}

func main() {
	service_url := "https://awspolicygen.s3.amazonaws.com/js/policies.js"
	//base_url := "https://docs.aws.amazon.com/service-authorization/latest/reference/"
	//service_json_url := base_url + "toc-contents.json"

	services_metadata, err := get_service_metadata(service_url)
	error_check(err)



	//awsglobalconditionkeys.csv
	for _, conditionKeys := range services_metadata.ConditionKeys{
		fmt.Println(conditionKeys)
	}

	//awsoperators.csv
	for _, operators := range services_metadata.ConditionOperators{
		fmt.Println(operators)
	}

	//awsmultivaluedprefix.csv follow up bc theres alot more here
	for _, details := range services_metadata.ServiceMap {
		fmt.Println( details.StringPrefix)
	}

	/*
		for _, data := range services_metadata.ServiceMap {
			for _,actions := range data.Actions{
				fmt.Println("\t",actions)
			}

		}
	
		
	for _, links := range get_services_json_href(service_json_url) {
		get_service_dict(base_url + links)
	}
		*/
		
}
