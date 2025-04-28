package genericresource

import (
	"testing"
)

func TestGenericResourceWithInt(t *testing.T) {
	gr := NewGenericResource[int](
		42,         // initialValue
		100,        // initialReadValue
		200,        // initialAcqRelValue
		50,         // initialAtomic
		60,         // initialMixed
		"test-int", // initialDesc
		"id-123",   // initialID
	)

	// Test proper locking
	gr.SetData(99, "updated-int")
	val, desc := gr.GetData()
	if val != 99 || desc != "updated-int" {
		t.Errorf("Expected (99, updated-int), got (%d, %s)", val, desc)
	}
}

func TestGenericResourceWithString(t *testing.T) {
	gr := NewGenericResource[string](
		"hello",       // initialValue
		"world",       // initialReadValue
		"locked",      // initialAcqRelValue
		50,            // initialAtomic
		60,            // initialMixed
		"test-string", // initialDesc
		"id-456",      // initialID
	)

	// Test proper locking
	gr.SetData("updated", "updated-string")
	val, desc := gr.GetData()
	if val != "updated" || desc != "updated-string" {
		t.Errorf("Expected (updated, updated-string), got (%s, %s)", val, desc)
	}
}
