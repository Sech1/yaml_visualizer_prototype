package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type JsonData struct {
	Name          string          `json:"name"`
	Included      bool            `json:"included"`
	ParentObjects []ParentObjects `json:"parentObjects"`
}

type ParentObjects struct {
	Name string
}

// Individual graph object as directed by jsMind.
type JsMindGraphObj struct {
	Id        string           `json:"id"`
	Topic     string           `json:"topic"`
	Direction string           `json:"direction"`
	Expanded  bool             `json:"expanded"`
	Children  []JsMindGraphObj `json:"children"`
}

// Master JSON object containing meta data and root node.
type JsMindJsonObj struct {
	Meta   MetaObj        `json:"meta"`
	Format string         `json:"format"`
	Data   JsMindGraphObj `json:"data"`
}

// Struct to hold misc. metadata.
type MetaObj struct {
	Name    string `json:"name"`
	Author  string `json:"author"`
	Version string `json:"version"`
}

// Yaml data obj to tie directories to their kustomization files.
type YamlDataObj struct {
	YamlPath string
	YamlName string
	YamlObj  interface{}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var debug = false
	if os.Args[1] == "-d" {
		debug = true
	} else if os.Args[1] == "-p" {
		debug = false
	}

	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Static("/index", "./public")

	// Get current working directory
	myDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	// Get route for API endpoint to return json data to build mindmap
	router.GET("/api", func(c *gin.Context) {
		// Return list of all yaml file locations
		yamlFiles, err := WalkMatch(fmt.Sprintf("%s/public/yaml/", myDir), "kustomization.yaml")
		fmt.Println(err)

		// Read all yaml files into an array
		var yamlArray []YamlDataObj
		for i, v := range yamlFiles {
			yamlArray = ReadYaml(v, yamlArray)
			//yamlArray = ReadYaml("pikachu", yamlArray)
			fmt.Println(yamlArray[i])
			fmt.Println("boi")
		}
		msg := ReturnGraphJson(yamlArray)
		var test = string(msg)
		log.Printf(test)
		c.JSON(http.StatusOK, msg)
	})

	if debug {
		router.Run(":8080")
	} else {
		router.RunTLS(":8080", "/server_keys/cert.pem", "/server_keys/privkey.pem")
	}
}

// Function that handles turning the yaml interfaces into JSON data and creating the required relationships.
func ReturnGraphJson(yamlArray []YamlDataObj) []byte {
	var children []JsMindGraphObj
	for i, v := range yamlArray {
		log.Printf("Processing Yaml # %d", i)
		log.Printf(" [x] Pulled JSON: %s", v)
		//test := .(map[string]interface{})["metadata"].(map[string]interface{})["name"]
		if v.YamlObj != nil {
			if val, ok := v.YamlObj.(map[string]interface{})["resources"]; ok {
				var newNode JsMindGraphObj
				if valObj, ok := val.([]interface{}); ok {
					newNode.Id = v.YamlName
					newNode.Expanded = true
					newNode.Direction = "right"
					newNode.Topic = v.YamlName
					newNode.Children = []JsMindGraphObj{}
					for _, vk := range valObj {
						var holder=vk.(string)
						 holder2 := strings.SplitN(holder, "/", -1) 
						//var startingNode JsMindGraphObj
						//vk2 is an array of the string vk.(string) split up buy the "/"
						//this loop is suposed to add each word file name from vk.(string)
						//to tempObj.Children
						for _,vk2 :=range holder2{
						
						var tempObj = JsMindGraphObj{Id: vk2, Topic: vk2,
							Direction: "right", Expanded: true, Children: []JsMindGraphObj{}}
						//im making a new item to add to tempObj.Children
						// this item means nothing and is only a placeholder for testing
						var tempObj2 = JsMindGraphObj{Id: "fake", Topic: "fake",
							Direction: "right", Expanded: true, Children: []JsMindGraphObj{}}
						//THIS IS THE LINE
						tempObj.Children = append(tempObj.Children, tempObj2)
						newNode.Children = append(newNode.Children, tempObj)
						
						}
						
						
					}

					
					 
					children = append(children, newNode)
					log.Printf(valObj[0].(string))
				}
			}
		}
	}


	

	jsonRoot := JsMindJsonObj{Meta: MetaObj{Name: "jsMindYaml", Author: "yaml", Version: "1.0"}, Format: "node_tree",
		Data: JsMindGraphObj{Id: "root", Topic: "Yaml Root Directory", Direction: "", Expanded: true,
			Children: children}}
	returnArray, err := json.Marshal(jsonRoot)
	if err != nil {
		log.Println(err)
	}
	return returnArray
}

// Function that handles turning the yaml interfaces into JSON data and creating the required relationships.
func ReturnGraphJson2(yamlArray []YamlDataObj) []byte {
	var children []JsMindGraphObj
	for i, v := range yamlArray {
		log.Printf("Processing Yaml # %d", i)
		log.Printf(" [x] Pulled JSON: %s", v)
		//test := .(map[string]interface{})["metadata"].(map[string]interface{})["name"]
		if v.YamlObj != nil {
			if val, ok := v.YamlObj.(map[string]interface{})["resources"]; ok {
				var newNode JsMindGraphObj
				if valObj, ok := val.([]interface{}); ok {
					newNode.Id = "boi"
					newNode.Expanded = true
					newNode.Direction = "right"
					newNode.Topic = v.YamlName
					newNode.Children = []JsMindGraphObj{}
					for _, vk := range valObj {
						var tempObj = JsMindGraphObj{Id: vk.(string), Topic: vk.(string),
							Direction: "", Expanded: true, Children: []JsMindGraphObj{}}
						newNode.Children = append(newNode.Children, tempObj)
					}
					children = append(children, newNode)
					log.Printf(valObj[0].(string))
				}
			}
		}
	}


	

	jsonRoot := JsMindJsonObj{Meta: MetaObj{Name: "jsMindYaml", Author: "yaml", Version: "1.0"}, Format: "node_tree",
		Data: JsMindGraphObj{Id: "root", Topic: "Yaml Root Directory", Direction: "", Expanded: true,
			Children: children}}
	returnArray, err := json.Marshal(jsonRoot)
	if err != nil {
		log.Println(err)
	}
	return returnArray
}
//
func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func readFile(filePath string, yaml *[]string) {
	strings.Replace(filePath, "/", "hello there", -1)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	
	*yaml = append(*yaml, string(data))
	
}

// Handles reading yaml files and putting the data into interfaces
func ReadYaml(filePath string, yamlArray []YamlDataObj) []YamlDataObj {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	var body interface{}
	if err := yaml.Unmarshal(yamlFile, &body); err != nil {
		panic(err)
	}
	body = convert(body)
	if b, err := json.Marshal(body); err != nil {
		panic(err)
	} else {
		fmt.Printf("Output: %s\n", b)
	}
	myDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	yamlName := strings.Replace(filePath,myDir, "", -1)
	yamlName = strings.Replace(yamlName, "yaml", "hello there", -1)
	yamlObj := YamlDataObj{filePath, yamlName, body}
	yamlArray = append(yamlArray, yamlObj)
	return yamlArray
}

// Recursive helper function for ReadYaml
// Creates interfaces objects of unknown depth.
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
