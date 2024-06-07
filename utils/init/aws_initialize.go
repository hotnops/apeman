package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

func get_start_index(array[] string) int{
	index := 0
	for i, words := range array{
		if words == "\"serviceMap\""{
			index = i
		}
	}
	return index
}




func main() {


	SERVICE_URL := "https://awspolicygen.s3.amazonaws.com/js/policies.js"
  //url_base := "https://docs.aws.amazon.com/service-authorization/latest/reference"	
	 
	resp, err := http.Get(SERVICE_URL)
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		 log.Fatalln(err)
	}
	//Convert the body to type string
	array := regexp.MustCompile("[\\:\\,\\.\\{\\}\\]]").Split(string(body), -1)

	start_index := get_start_index(array)

	for i := start_index; i < len(array); i++{
		fmt.Println(i, " => ", string(array[i]))
	}
}