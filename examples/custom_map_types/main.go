package main

import (
	"fmt"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// Custom type alias (like domain.MapOfAny)
type MapOfAny map[string]any

func main() {
	fmt.Println("liquidgo Custom Map Type Fix Demonstration")
	fmt.Println("==========================================")
	fmt.Println()

	// Test 1: Direct map[string]interface{} - WORKS ✅
	data1 := map[string]interface{}{
		"workspace": map[string]interface{}{
			"id": "test-123",
		},
	}

	// Test 2: Custom type MapOfAny - NOW WORKS ✅
	workspaceData := MapOfAny{
		"id": "test-456",
	}
	data2 := map[string]interface{}{
		"workspace": workspaceData,
	}

	// Test 3: After conversion to map[string]interface{} - WORKS ✅
	workspaceConverted := make(map[string]interface{})
	for k, v := range workspaceData {
		workspaceConverted[k] = v
	}
	data3 := map[string]interface{}{
		"workspace": workspaceConverted,
	}

	template := "{{ workspace.id }}"

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, _ := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})

	result1 := tmpl.Render(data1, nil)
	result2 := tmpl.Render(data2, nil)
	result3 := tmpl.Render(data3, nil)

	fmt.Println("Test 1 - map[string]interface{}:")
	fmt.Printf("  Template: %s\n", template)
	fmt.Printf("  Result:   %q\n", result1)
	fmt.Printf("  Status:   %s\n\n", checkResult(result1, "test-123"))

	fmt.Println("Test 2 - Custom type MapOfAny:")
	fmt.Printf("  Template: %s\n", template)
	fmt.Printf("  Result:   %q\n", result2)
	fmt.Printf("  Status:   %s\n\n", checkResult(result2, "test-456"))

	fmt.Println("Test 3 - MapOfAny after conversion:")
	fmt.Printf("  Template: %s\n", template)
	fmt.Printf("  Result:   %q\n", result3)
	fmt.Printf("  Status:   %s\n\n", checkResult(result3, "test-456"))

	// Additional demonstration: nested custom types
	fmt.Println("Additional Test - Nested Custom Types:")
	nestedData := map[string]interface{}{
		"company": MapOfAny{
			"name": "Acme Corp",
			"workspace": MapOfAny{
				"id":   "ws-789",
				"name": "Production",
			},
		},
	}

	nestedTemplate := "{{ company.name }}: {{ company.workspace.name }} ({{ company.workspace.id }})"
	nestedTmpl, _ := liquid.ParseTemplate(nestedTemplate, &liquid.TemplateOptions{Environment: env})
	nestedResult := nestedTmpl.Render(nestedData, nil)

	fmt.Printf("  Template: %s\n", nestedTemplate)
	fmt.Printf("  Result:   %q\n", nestedResult)
	fmt.Printf("  Expected: %q\n", "Acme Corp: Production (ws-789)")
	fmt.Printf("  Status:   %s\n", checkResult(nestedResult, "Acme Corp: Production (ws-789)"))
}

func checkResult(got, expected string) string {
	if got == expected {
		return "✅ PASS"
	}
	return "❌ FAIL"
}
