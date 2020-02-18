package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type BindFile struct {
	Name  string                `form:"name" binding:"required"`
	Email string                `form:"email" binding:"required"`
	File  *multipart.FileHeader `form:"file" binding:"required"`
}

type JsonData struct {
	Name          string          `json:"name"`
	Included      bool            `json:"included"`
	ParentObjects []ParentObjects `json:"parentObjects"`
}

type ParentObjects struct {
	Name string
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

	router.GET("/api", func(c *gin.Context) {
		// Return list of all yaml file locations
		yamlFiles, err := WalkMatch(fmt.Sprintf("%s/public/yaml/", myDir), "*.yaml")
		fmt.Println(err)

		// Read all yaml files into an array
		var yamlArray []interface{}
		for i, v := range yamlFiles {
			yamlArray = ReadYaml(v, yamlArray)
			fmt.Println(yamlArray[i])
		}
		msg := ReturnGraphJson(yamlArray)
		c.JSON(http.StatusOK, msg)
	})

	router.POST("/upload", func(c *gin.Context) {
		var bindFile BindFile

		// Bind file
		if err := c.ShouldBind(&bindFile); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
			return
		}

		// Save uploaded file
		file := bindFile.File
		dst := filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully with fields name=%s and email=%s.", file.Filename, bindFile.Name, bindFile.Email))
	})
	if debug {
		router.Run(":8080")
	} else {
		router.RunTLS(":8080", "/server_keys/cert.pem", "/server_keys/privkey.pem")
	}
}

func ReturnGraphJson(yamlArray []interface{}) []byte {
	var jsonData []JsonData
	for i, v := range yamlArray {
		log.Printf("Processing Yaml # %d", i)
		log.Printf(" [x] Pulled JSON: %s", v)
		//test := .(map[string]interface{})["metadata"].(map[string]interface{})["name"]
		if v != nil {
			if val, ok := v.(map[string]interface{})["metadata"]; ok {
				var newNode JsonData
				if valObj, ok := val.(map[string]interface{}); ok {
					newNode.Name = valObj["name"].(string)
					newNode.Included = true
					newNode.ParentObjects = []ParentObjects{}
					jsonData = append(jsonData, newNode)
				}
			}
		}
	}
	returnArray, err := json.Marshal(jsonData)
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

func ReadYaml(filePath string, yamlArray []interface{}) []interface{} {
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
	yamlArray = append(yamlArray, body)
	return yamlArray
}

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
