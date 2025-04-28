package genericresource

import (
	"sync"
)

// NonGenericResource is identical to GenericResource but without generics.
type NonGenericResource struct {
	mu sync.Mutex
	// +checklocks:mu
	value int
	// +checklocks:mu
	description string

	id string // This field is not guarded by mu

	rwMu sync.RWMutex
	// +checklocks:rwMu
	readGuardedValue int

	// +checkatomic
	atomicValue int32

	// +checkatomic
	// +checklocks:mu
	mixedValue int32

	acquireReleaseMu sync.Mutex
	// +checklocks:acquireReleaseMu
	acquireReleaseValue int
}

// NewNonGenericResource creates a new NonGenericResource.
func NewNonGenericResource(initialValue, initialReadValue, initialAcqRelValue int, initialAtomic, initialMixed int32, initialDesc, initialID string) *NonGenericResource {
	return &NonGenericResource{
		value:               initialValue,
		description:         initialDesc,
		id:                  initialID,
		readGuardedValue:    initialReadValue,
		atomicValue:         initialAtomic,
		mixedValue:          initialMixed,
		acquireReleaseValue: initialAcqRelValue,
	}
}

// SetData correctly locks the mutex before writing to the guarded fields.
func (ngr *NonGenericResource) SetData(val int, desc string) {
	ngr.mu.Lock()
	ngr.value = val
	ngr.description = desc
	ngr.mu.Unlock()
}

// GetData correctly locks the mutex before reading the guarded fields.
func (ngr *NonGenericResource) GetData() (int, string) {
	ngr.mu.Lock()
	v := ngr.value
	d := ngr.description
	ngr.mu.Unlock()
	return v, d
}

// setDataLocked sets the guarded values, assuming the lock is already held by the caller.
// The +checklocks annotation enforces this assumption.
// +checklocks:ngr.mu
func (ngr *NonGenericResource) setDataLocked(val int, desc string) {
	ngr.value = val
	ngr.description = desc
}

// SetDataWithHelper demonstrates calling an annotated function correctly (lock held).
func (ngr *NonGenericResource) SetDataWithHelper(val int, desc string) {
	ngr.mu.Lock()
	ngr.setDataLocked(val, desc) // Correct: Lock 'ngr.mu' is held.
	ngr.mu.Unlock()
}
