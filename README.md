**This project uses:**
<br>
**[JsMind](https://github.com/hizzgdev/jsmind) to build the mindmap in javascript**<br>
**[Gin](https://github.com/gin-gonic/gin) for the webserver back-end**<br>
**[yaml.v2](https://github.com/go-yaml/yaml) for golang yaml support**<br>

---
[A working example of this project is hosted at https://graph.sech.dev](https://graph.sech.dev)
---
#### This assumes you have Go Installed 
##### You also need Gin Gonic & yaml.v2 installed properly as a go packages

Instructions to install Gin-Gonic can be found [here](https://github.com/gin-gonic/gin)<br>
Instructions to install yaml.v2 can be found [here](https://gopkg.in/yaml.v2)
> go get -u github.com/gin-gonic/gin <br>
> go get gopkg.in/yaml.v2

This project should be put in $GOPATH/src
* Ex: Go/src/yaml_visualizer_prototype
> cd ~/Go/src <br>
> git clone GITREPOLINK <br>
> cd yaml_visualizer_prototype <br>
> go run main.go <br>

 To run the graph server in debug mode type:
 > go run main.go -d 

To run the graph server in production mode type:
> go run main.go -p
 
