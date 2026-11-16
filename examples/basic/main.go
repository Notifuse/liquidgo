package main

import (
	"context"
	"fmt"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

func main() {
	// Basic template rendering
	tmpl, err := liquid.ParseTemplate("Hello {{ name }}!", nil)
	if err != nil {
		panic(err)
	}

	res, err := tmpl.Render(context.Background(), map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Output) // Output: Hello World!
}
