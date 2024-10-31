package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

func main() {
	searchKey := flag.String("key", "", "Key to search for (in all fields)")
	searchValue := flag.String("value", "", "Value to search for (in all fields, scalar values only)")
	filename := flag.String("file", "", "File to search (JSON or YAML)")

	flag.Parse()

	if *searchKey == "" && *searchValue == "" {
		fmt.Println("Error: Must specify either -key or -value")
		os.Exit(1)
	}

	var data []byte
	var err error

	if *filename != "" {
		data, err = os.ReadFile(*filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
	} else {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			os.Exit(1)
		}
	}

	var res gjson.Result
	if strings.HasSuffix(strings.ToLower(*filename), ".yaml") || strings.HasSuffix(strings.ToLower(*filename), ".yml") {
		var yamlData interface{}
		err = yaml.Unmarshal(data, &yamlData)
		if err != nil {
			fmt.Println("Error parsing YAML:", err)
			os.Exit(1)
		}
		jsonData, err := json.Marshal(yamlData)
		if err != nil {
			fmt.Println("Error converting YAML to JSON:", err)
			os.Exit(1)
		}
		res = gjson.ParseBytes(jsonData)
	} else {
		res = gjson.ParseBytes(data)
	}

	if *searchKey != "" {
		searchAllKeys(res, *searchKey)
	}
	if *searchValue != "" {
		searchAllValues(res, *searchValue)
	}
}

func searchAllKeys(res gjson.Result, searchKey string) {
	searchRecursiveKeys(res, searchKey, "")
}

func searchAllValues(res gjson.Result, searchValue string) {
	searchRecursiveValues(res, searchValue, "")
}

func searchRecursiveKeys(res gjson.Result, searchKey, currentPath string) {
	if res.IsObject() {
		for key, value := range res.Map() {
			newPath := buildPath(currentPath, key)
			if strings.Contains(key, searchKey) {
				fmt.Printf("Path: %s\nValue: %s\n\n", newPath, value.String())
			}
			searchRecursiveKeys(value, searchKey, newPath)
		}
	} else if res.IsArray() {
		for i, value := range res.Array() {
			newPath := buildPath(currentPath, fmt.Sprintf("[%d]", i))
			searchRecursiveKeys(value, searchKey, newPath)
		}
	}
}

func searchRecursiveValues(res gjson.Result, searchValue, currentPath string) {
	if res.IsObject() {
		for key, value := range res.Map() {
			newPath := buildPath(currentPath, key)
			searchRecursiveValues(value, searchValue, newPath)
		}
	} else if res.IsArray() {
		for i, value := range res.Array() {
			newPath := buildPath(currentPath, fmt.Sprintf("[%d]", i))
			searchRecursiveValues(value, searchValue, newPath)
		}
	} else if res.Type.String() != "JSON" && res.Type.String() != "array" && res.Type.String() != "object" {
		if strings.Contains(res.String(), searchValue) {
			fmt.Printf("Path: %s\nValue: %s\n\n", currentPath, res.String())
		}
	}
}

func buildPath(currentPath, key string) string {
	if currentPath == "" {
		return key
	}
	return currentPath + "." + key
}
