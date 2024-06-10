package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
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

	func get_service_metadata(url string) map[string]ServiceDetails{

		response, err := http.Get(url)
		if err != nil {
			log.Fatal("Error fetching data:", err)
		}
		defer response.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Error reading response body:", err)
		}

		// Define the structure to store the JSON data
		var config PolicyEditorConfig

		// Unmarshal the JSON data into the struct
		output := regexp.MustCompile("app.PolicyEditorConfig=")
		edit := output.ReplaceAllString(string(body), "")

		//fmt.Println(edit)

		err = json.Unmarshal([]byte(edit), &config)
		if err != nil {
			log.Fatal("Error parsing JSON:", err)
		}

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
			return config.ServiceMap
	}

	func write_data_to_csv(){
		//output this all to a csv

	}

	func main() {
		url := "https://awspolicygen.s3.amazonaws.com/js/policies.js"

		fmt.Println(get_service_metadata(url))
	}