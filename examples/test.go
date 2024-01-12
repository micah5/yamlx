package main

import (
	"fmt"
	"github.com/micah5/yamlx"
)

type Server struct {
	Name   string `yamlx:"name"`
	Host   string `yamlx:"host"`
	Port   int    `yamlx:"port"`
	Engine string `yamlx:"engine,omitempty"`
}

type Config struct {
	Servers []Server `yamlx:"servers"`
}

func main() {
	data := `
environments: &environments
  - sandbox
  - dev
  - staging
  - prod

db_settings: &db_settings
  engine:
    latest: mysql8.2
    stable: mysql5.7

subdomains: &subdomains [example, myapp]

servers:
  !for idx, name in *environments:
    - name: *name
      host: ${name}.${join(".", subdomains)}.com
      port: 22
    - name: ${name}-db
      host: 192.168.1.${(idx + 1) * 100}
      port: 3306
      engine: ${name == "prod" ? db_settings.engine.stable : db_settings.engine.latest}
`
	var config Config
	err := yamlx.Unmarshal([]byte(data), &config)
	if err != nil {
		panic(err)
	}

	// convert to yaml for demonstration purposes
	yamlOutput, _ := yamlx.Marshal(config)
	fmt.Println(string(yamlOutput))
}
