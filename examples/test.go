package main

import (
	"fmt"
	"github.com/micah5/yamlx"
)

type Person struct {
	Name    string `yamlx:"first_name"`
	Surname string `yamlx:"last_name"`
	Age     int    `yamlx:"age"`
}

type Organization struct {
	Name   string   `yamlx:"name"`
	People []Person `yamlx:"people"`
}

func main() {
	data := `
# Define variables
variables:
  environment: "production"
  server_names: ["server1", "server2", "server3"]

# Define a function
functions:
  generate_ip:
    parameters: [name]
    body: |
      return "192.168.1." + str(ord(name[-1]))

# Main configuration
configuration:
  servers:
    - for_each: ${server_names}
      server:
        name: ${item}
        ip: ${generate_ip(item)}
        environment: ${environment}
        status: "${if environment == 'production' then 'active' else 'inactive'}"
`
	var org Organization
	yamlx.Unmarshal([]byte(data), &org)

	fmt.Println(org)
}
