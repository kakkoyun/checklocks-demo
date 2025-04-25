package main

import (
	"fmt"

	"github.com/kakkoyun/checklocks-demo/pkg/resource"
)

func main() {
	fmt.Println("Checklocks Demo")

	// Create a resource
	pr := resource.NewProtectedResource(100, 110, 120, 130, 140, "Initial Description", "ID-MAIN-001")

	// Get initial data (correctly locked)
	val, desc := pr.GetData()
	fmt.Printf("Initial Data: Value=%d, Description='%s'\n", val, desc)

	// Get initial ID (unguarded)
	id := pr.GetID()
	fmt.Printf("Initial ID: %s\n", id)

	// Set some data (correctly locked)
	pr.SetData(200, "Updated Description")
	val, desc = pr.GetData()
	fmt.Printf("Updated Data: Value=%d, Description='%s'\n", val, desc)

	// Set ID (unguarded)
	pr.SetID("ID-MAIN-UPDATED")
	id = pr.GetID()
	fmt.Printf("Updated ID: %s\n", id)

	// We don't call the Incorrect* methods here, as their violations
	// are checked by the linter and confirmed by the tests.
	fmt.Println("Demo finished.")
}
