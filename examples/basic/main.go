package main

import (
	"fmt"

	"github.com/Notifuse/liquidgo/liquid"
)

func main() {
	// Basic template rendering
	tmpl, err := liquid.ParseTemplate("Hello {{ name }}!", nil)
	if err != nil {
		panic(err)
	}

	output := tmpl.Render(map[string]interface{}{
		"name": "World",
	}, nil)

	fmt.Println(output) // Output: Hello World!
}
