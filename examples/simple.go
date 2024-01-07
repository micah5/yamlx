package main

import (
	"github.com/micah5/yamlx"
)

func main() {
	// Sample YAML data
	yamlData := `
key1: value1
key2:
	key2_1: &anchor1 value2_1
	key2_2:
		- value2_2_1
		- &anchor2 value2_2_2
		- key2_2_3: value2_2_3
		- 1
		- 1.1
		- true
		- a: b
		  c: d
		  e: f
	key2_3: &anchor23
		key2_3_1: value2_3_1
		key2_3_2: *anchor2
		key2_3_3: *anchor1
		int: 1
		float: 1.2
		bool: true
key3: &anchor3 value3
key4:
	- value4_1
	- value4_2
	- *anchor3
	- key4_4:
		yee: yee
key5:
	<<: *anchor23
	key5_1: value5_1
`
	yamlx.Parse(yamlData)
}
