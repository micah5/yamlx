package main

import (
	"github.com/micah5/yamlx"
)

func main() {
	// Sample YAML data
	yamlData := `
defaults:
	colonial: &colonial_window
		width: 0.5
		height: 1.0
		sill_type: 1 # this corresponds to an enum which could be elaborated within the docs. e.g. 0: no sill, 1: simple sill, 2: bevelled sill etc
		corbel_type: 1 # something similar to above
		num_corbels: 2
		outer_frame_thickness: 0.15
		inner_frame_thickness: 0.075
		blah: &blah2
			- 1
			- 2

floors:
	- walls:
		-	width: 0.5
			height: 1.0
			sill_type: 1
	- a : b
	- c: d
	- e: g
    	- f: *colonial_window
    	- blah3:
        	<<: *colonial_window
        	hello: there
`
	yamlx.Parse(yamlData)
}
