package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

	// Create list of yaml files & data
	yamlFiles, err := WalkMatch(fmt.Sprintf("%s/public/yaml/", myDir), "kustomization.yaml")
	// Read all yaml files into an array
	var yamlArray []YamlDataObj
	for i, v := range yamlFiles {
		yamlArray = ReadYaml(v, yamlArray)
		fmt.Println(yamlArray[i])
	}

	// Get route for API endpoint to return json data to build mindmap
	router.GET("/api", func(c *gin.Context) {
		// Return list of all yaml file locations
		fmt.Println(err)

		msg := ReturnGraphJson(yamlArray)
		var test = string(msg)
		log.Printf(test)
		c.JSON(http.StatusOK, msg)
	})

	router.GET("/yaml_links", func(c *gin.Context) {
		c.JSON(http.StatusOK, CreateNodeLinkDict(yamlArray, myDir))
	})

	if debug {
		router.Run(":8080")
	} else {
		router.RunTLS(":8080", "/server_keys/cert.pem", "/server_keys/privkey.pem")
	}
}

func CreateNodeLinkDict(yamlArray []YamlDataObj, directory string) []byte {
	linkMap := make(map[string]string)

	for _, v := range yamlArray {
		_, exist := linkMap[v.YamlName]
		if !exist {
			link := strings.Replace(v.YamlPath, directory, "", 1)
			link = strings.Replace(link, "public", "index", 1)
			originalLink := link
			linkMap[v.YamlName] = link

			if v.YamlObj != nil {
				if val, ok := v.YamlObj.(map[string]interface{})["resources"]; ok {
					if valObj, ok := val.([]interface{}); ok {
						for _, vk := range valObj {
							_, exist := linkMap[vk.(string)]
							if !exist {
								if strings.Contains(vk.(string), ".yaml") {
									link = fmt.Sprintf("%s/%s", strings.Replace(linkMap[v.YamlName],
										"/kustomization.yaml", "", 1),
										strings.Replace(vk.(string), "./", "", 1))
								} else if strings.Contains(vk.(string), "../") {
									link = originalLink
									regEx := regexp.MustCompile(`\.\./`)
									matches := regEx.FindAllString(vk.(string), -1)
									splitLink := strings.Split(strings.Replace(originalLink, "/kustomization.yaml", "", 1), "/")
									directoryWalkCount := len(splitLink) - len(matches)
									var newLink = strings.Replace(v.YamlPath, directory, "", 1)
									newLink = strings.Replace(newLink, "public", "index", 1)
									var sbLink strings.Builder
									for i := 1; i < directoryWalkCount; i++ {
										sbLink.WriteString(fmt.Sprintf("/%s", splitLink[i]))
									}
									sbLink.WriteString("/")
									sbLink.WriteString(strings.Replace(vk.(string), "../", "", len(matches)))
									sbLink.WriteString("/kustomization.yaml")
									link = sbLink.String()
								} else if strings.Contains(vk.(string), "./") {
									var sbLink strings.Builder
									sbLink.WriteString(fmt.Sprintf("%s/%s", strings.Replace(linkMap[v.YamlName],
										"/kustomization.yaml", "", 1),
										strings.Replace(vk.(string), "./", "", 1)))
									sbLink.WriteString("/kustomization.yaml")
									link = sbLink.String()
								} else {
									link = fmt.Sprintf("%s/%s", strings.Replace(linkMap[v.YamlName],
										"/kustomization.yaml", "", 1),
										strings.Replace(vk.(string), "./", "", 1))
								}
								linkMap[vk.(string)] = link
							}
						}
					}
				}
			}
		}
	}

	returnData, _ := json.Marshal(linkMap)
	return returnData
}

// Function that handles turning the yaml interfaces into JSON data and creating the required relationships.
func ReturnGraphJson(yamlArray []YamlDataObj) []byte {
	var direction = "right"
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
					newNode.Expanded = false
					newNode.Direction = direction
					if direction == "right" {
						direction = "left"
					} else {
						direction = "right"
					}
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
	yamlName := strings.Replace(filePath, fmt.Sprintf("%s/public/yaml/", myDir), "", 1)
	yamlName = strings.Replace(yamlName, "/kustomization.yaml", "", 1)
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
