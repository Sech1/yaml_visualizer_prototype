package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"opendev.org/airship/airshipctl/pkg/document"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

type Request struct {
	RequestedId string
}

func main() {
	// This map is used to keep track of node id's and the Json Data that correlates to them.
	linkHashMap := make(map[string][]uint8)
	var debug bool
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

		msg := MakeGraph(yamlArray, &linkHashMap)
		var test = string(msg)
		log.Printf(test)
		c.JSON(http.StatusOK, msg)
	})

	router.GET("/yaml_links", func(c *gin.Context) {
		var request = c.Request.URL.Query()
		c.JSON(http.StatusOK, linkHashMap[request.Get("requestedId")])
	})

	if debug {
		router.Run(":8080")
	} else {
		router.RunTLS(":8080", "/server_keys/cert.pem", "/server_keys/privkey.pem")
	}
}

// Handles looping through yamlArray to create links to yaml documents
// Used by API to enable view yaml on-click functionality.
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

func MakeGraph(yamlArray []YamlDataObj, uniqueIdArray *map[string][]uint8) []byte {
	// Map of string that correlates to an array of documents
	nameSpaceMap := make(map[string][]document.Document)
	var direction = "right"
	var children []JsMindGraphObj
	var folder []document.Document

	//var namespaces []NameSpace
	for i := range yamlArray {
		docBundle, err := document.NewBundleByPath(filepath.Dir(yamlArray[i].YamlPath))
		log.Printf("Processing Yaml # %d", i)
		if err != nil {
			log.Println(err)
		} else {
			folder, err = docBundle.GetAllDocuments()
			if folder != nil {
				for _, d := range folder {
					nameSpaceMap[d.GetNamespace()] = append(nameSpaceMap[d.GetNamespace()], d)
				}
			}
		}
	}

	for nameSpace, value := range nameSpaceMap {
		var name string
		var id string
		if nameSpace == "" {
			name = "No Namespace"
			id = GenerateNodeId(uniqueIdArray, "NoNamespace")
		} else {
			name = nameSpace
			id = GenerateNodeId(uniqueIdArray, nameSpace)
		}

		var newNode = CreateGenericNode(id, name, direction, uniqueIdArray, nil)

		nodeKindMap := make(map[string][]document.Document)
		for _, resource := range value {
			nodeKindMap[resource.GetKind()] = append(nodeKindMap[resource.GetKind()], resource)
		}
		for kindValue, kind := range nodeKindMap {
			var newKindNode = CreateGenericNode(kindValue, kindValue, direction, uniqueIdArray, nil)
			for _, documentObj := range kind {
				var newDocumentNode = CreateGenericNode(documentObj.GetName(), documentObj.GetName(), direction,
					uniqueIdArray, documentObj)
				var yamlData, _ = documentObj.MarshalJSON()
				log.Printf(string(yamlData))
				newKindNode.Children = append(newKindNode.Children, newDocumentNode)
			}
			newNode.Children = append(newNode.Children, newKindNode)
		}

		children = append(children, newNode)
		if direction == "right" {
			direction = "left"
		} else {
			direction = "right"
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

// Generates generic nodes used inside the json graph object.
func CreateGenericNode(id string, name string, direction string, uniqueIdArray *map[string][]uint8, documentObj document.Document) JsMindGraphObj {
	var newNode JsMindGraphObj
	newNode.Id = GenerateNodeId(uniqueIdArray, id)
	newNode.Topic = name
	newNode.Expanded = false
	newNode.Direction = direction
	newNode.Children = []JsMindGraphObj{}

	// If documentObj isn't nil - this node has a .yaml associated with it.
	// Find the set the id index to the JSON representation of that yaml.
	if documentObj != nil {
		(*uniqueIdArray)[newNode.Id], _ = documentObj.MarshalJSON()
		// If documentObj is nil , this node has no yaml associated with it
	} else {
		(*uniqueIdArray)[newNode.Id] = nil
	}

	return newNode
}

// This is used to check if an id has been used before, if it has it will create a new unique id.
// If it hasn't it returns the id passed in.
func GenerateNodeId(uniqueIdArray *map[string][]uint8, id string) string {
	// Use the Find function to see if an id exists.
	_, idUsed := Find(uniqueIdArray, id)
	if idUsed {
		// If it does - create unique id
		id = CreateUnusedId(uniqueIdArray, id)
	}
	// If it doesn't, return Id passed in.
	return id
}

// This function is used to walk the directory structure and find all files of a given pattern & return
// them inside an array of strings.
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

// Function to check if a string exists in an array
// Used to check if a unique Id has been used before.
func Find(slice *map[string][]uint8, val string) (int, bool) {
	// If key exists in slice return true
	if _, ok := (*slice)[val]; ok {
		return 1, true
	} else {
		// If it doesn't exist return false
		return 0, false
	}
}

// This is used to create a string when an Id is used
func CreateUnusedId(slice *map[string][]uint8, id string) string {
	idExists := true
	newId := ""
	var val = 0
	// Loop until an id value is incremented that does not currently exist in the map.
	for {
		newId = fmt.Sprintf("%s%d", id, val)
		_, idExists = Find(slice, newId)
		// Break if unique id has been found
		if !idExists {
			break
		}
		val++
	}
	// Once unique id is found return
	return newId
}
