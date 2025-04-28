package genericresource

import (
	"sync"
	"sync/atomic"

	"github.com/trailofbits/go-mutexasserts"
)

// GenericResource demonstrates a resource with some fields guarded by a mutex.
// This version uses generics to see if checklocks works with generic types.
type GenericResource[T any] struct {
	mu sync.Mutex
	// +checklocks:mu
	value T
	// +checklocks:mu
	description string

	id string // This field is not guarded by mu

	rwMu sync.RWMutex
	// +checklocks:rwMu
	readGuardedValue T

	// +checkatomic
	atomicValue int32

	// +checkatomic
	// +checklocks:mu
	mixedValue int32

	acquireReleaseMu sync.Mutex
	// +checklocks:acquireReleaseMu
	acquireReleaseValue T
}

// NewGenericResource creates a new GenericResource.
func NewGenericResource[T any](initialValue, initialReadValue, initialAcqRelValue T, initialAtomic, initialMixed int32, initialDesc, initialID string) *GenericResource[T] {
	return &GenericResource[T]{
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
func (gr *GenericResource[T]) SetData(val T, desc string) {
	gr.mu.Lock()
	gr.value = val
	gr.description = desc
	gr.mu.Unlock()
}

// GetData correctly locks the mutex before reading the guarded fields.
func (gr *GenericResource[T]) GetData() (T, string) {
	gr.mu.Lock()
	v := gr.value
	d := gr.description
	gr.mu.Unlock()
	return v, d
}

// setDataLocked sets the guarded values, assuming the lock is already held by the caller.
// The +checklocks annotation enforces this assumption.
// +checklocks:gr.mu
func (gr *GenericResource[T]) setDataLocked(val T, desc string) {
	gr.value = val
	gr.description = desc
}

// SetDataWithHelper demonstrates calling an annotated function correctly (lock held).
func (gr *GenericResource[T]) SetDataWithHelper(val T, desc string) {
	gr.mu.Lock()
	gr.setDataLocked(val, desc) // Correct: Lock 'gr.mu' is held.
	gr.mu.Unlock()
}

// GetID reads the unguarded ID field. No lock is needed.
func (gr *GenericResource[T]) GetID() string {
	return gr.id // Correct: No lock needed for unguarded field.
}

// GetReadGuardedValueCorrect correctly acquires the read lock.
func (gr *GenericResource[T]) GetReadGuardedValueCorrect() T {
	gr.rwMu.RLock()
	v := gr.readGuardedValue
	gr.rwMu.RUnlock()
	return v
}

// readDataRLocked requires the caller to hold at least the read lock.
// +checklocksread:gr.rwMu
func (gr *GenericResource[T]) readDataRLocked() T {
	return gr.readGuardedValue // Correct: Read lock is assumed held by annotation.
}

// CallReadDataRLockedCorrect calls an annotated function correctly (RLock held).
func (gr *GenericResource[T]) CallReadDataRLockedCorrect() T {
	gr.rwMu.RLock()
	v := gr.readDataRLocked() // Correct: Lock 'gr.rwMu' is read-held.
	gr.rwMu.RUnlock()
	return v
}

// ReadAtomicCorrect uses atomic operations on an atomic-only field.
func (gr *GenericResource[T]) ReadAtomicCorrect() int32 {
	return atomic.LoadInt32(&gr.atomicValue) // Correct: Atomic operation on atomic field.
}

// WriteMixedCorrect writes a mixed field atomically with the lock held (required for writes).
func (gr *GenericResource[T]) WriteMixedCorrect(v int32) {
	gr.mu.Lock()
	atomic.StoreInt32(&gr.mixedValue, v) // Correct: Lock is held and write is atomic.
	gr.mu.Unlock()
}

// AcquireAndSet acquires the lock and sets the value.
// +checklocksacquire:gr.acquireReleaseMu
func (gr *GenericResource[T]) AcquireAndSet(v T) {
	// Annotation requires lock NOT be held on entry.
	gr.acquireReleaseMu.Lock() // Acquires the lock.
	gr.acquireReleaseValue = v
	// Annotation implies lock IS held on exit.
}

// GetAndRelease reads the value and releases the lock.
// +checklocksrelease:gr.acquireReleaseMu
func (gr *GenericResource[T]) GetAndRelease() T {
	// Annotation requires lock BE held on entry.
	v := gr.acquireReleaseValue
	gr.acquireReleaseMu.Unlock() // Releases the lock.
	// Annotation implies lock IS NOT held on exit.
	return v
}

// CallAcquireReleaseCorrect demonstrates the correct acquire/release cycle.
func (gr *GenericResource[T]) CallAcquireReleaseCorrect() T {
	// Lock is not held here.
	var zeroVal T
	gr.AcquireAndSet(zeroVal) // Acquires lock.
	// Lock is now held.
	v := gr.GetAndRelease() // Releases lock.
	// Lock is not held here.
	return v
}

// FunctionToIgnore contains a violation but the analyzer is told to ignore it.
// +checklocksignore
func (gr *GenericResource[T]) FunctionToIgnore(v T) {
	gr.value = v // Violation: Access to gr.value without holding gr.mu.
	// This violation should NOT be reported by the linter due to +checklocksignore.
}

// helperCalledUnderLock is intended to ONLY be called when gr.mu is held.
// We use +checklocksignore because the analyzer can't know this context,
// but we guarantee it externally.
// +checklocksignore
func (gr *GenericResource[T]) helperCalledUnderLock(v T) {
	mutexasserts.AssertMutexLocked(&gr.mu)
	// This direct access would normally be a violation, but the function
	// is ignored by the analyzer.
	gr.value = v
}

// CallHelperUnderLockCorrectly demonstrates calling the ignored helper correctly.
func (gr *GenericResource[T]) CallHelperUnderLockCorrectly(v T) {
	gr.mu.Lock()
	gr.helperCalledUnderLock(v)
	gr.mu.Unlock()
}
