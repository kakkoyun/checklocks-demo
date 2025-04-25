package resource

import (
	"sync"
	"sync/atomic"
)

// ProtectedResource demonstrates a resource with some fields guarded by a mutex.
type ProtectedResource struct {
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

// NewProtectedResource creates a new ProtectedResource.
func NewProtectedResource(initialValue, initialReadValue, initialAcqRelValue int, initialAtomic, initialMixed int32, initialDesc, initialID string) *ProtectedResource {
	return &ProtectedResource{
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
func (pr *ProtectedResource) SetData(val int, desc string) {
	pr.mu.Lock()
	pr.value = val
	pr.description = desc
	pr.mu.Unlock()
}

// IncorrectSetData incorrectly writes to the guarded fields without locking.
// This should be flagged by the checklocks analyzer.
func (pr *ProtectedResource) IncorrectSetData(val int, desc string) {
	pr.value = val        // Error: Lock 'pr.mu' is not held for pr.value
	pr.description = desc // Error: Lock 'pr.mu' is not held for pr.description
}

// GetData correctly locks the mutex before reading the guarded fields.
func (pr *ProtectedResource) GetData() (int, string) {
	pr.mu.Lock()
	v := pr.value
	d := pr.description
	pr.mu.Unlock()
	return v, d
}

// setDataLocked sets the guarded values, assuming the lock is already held by the caller.
// The +checklocks annotation enforces this assumption.
// +checklocks:pr.mu
func (pr *ProtectedResource) setDataLocked(val int, desc string) {
	pr.value = val
	pr.description = desc
}

// SetDataWithHelper demonstrates calling an annotated function correctly (lock held).
func (pr *ProtectedResource) SetDataWithHelper(val int, desc string) {
	pr.mu.Lock()
	pr.setDataLocked(val, desc) // Correct: Lock 'pr.mu' is held.
	pr.mu.Unlock()
}

// IncorrectSetDataWithHelper demonstrates calling an annotated function incorrectly (lock not held).
// This should be flagged by the checklocks analyzer.
func (pr *ProtectedResource) IncorrectSetDataWithHelper(val int, desc string) {
	pr.setDataLocked(val, desc) // Error: Lock 'pr.mu' is not held before calling function requiring it.
}

// GetID reads the unguarded ID field. No lock is needed.
func (pr *ProtectedResource) GetID() string {
	return pr.id // Correct: No lock needed for unguarded field.
}

// SetID sets the unguarded ID field. No lock is needed.
func (pr *ProtectedResource) SetID(newID string) {
	pr.id = newID // Correct: No lock needed for unguarded field.
}

// --- RWMutex and Read Locks ---

// GetReadGuardedValueCorrect correctly acquires the read lock.
func (pr *ProtectedResource) GetReadGuardedValueCorrect() int {
	pr.rwMu.RLock()
	v := pr.readGuardedValue
	pr.rwMu.RUnlock()
	return v
}

// GetReadGuardedValueIncorrect accesses a field guarded by RWMutex without any lock.
func (pr *ProtectedResource) GetReadGuardedValueIncorrect() int {
	return pr.readGuardedValue // Error: Lock 'pr.rwMu' is not held.
}

// readDataRLocked requires the caller to hold at least the read lock.
// +checklocksread:pr.rwMu
func (pr *ProtectedResource) readDataRLocked() int {
	return pr.readGuardedValue // Correct: Read lock is assumed held by annotation.
}

// CallReadDataRLockedCorrect calls an annotated function correctly (RLock held).
func (pr *ProtectedResource) CallReadDataRLockedCorrect() int {
	pr.rwMu.RLock()
	v := pr.readDataRLocked() // Correct: Lock 'pr.rwMu' is read-held.
	pr.rwMu.RUnlock()
	return v
}

// CallReadDataRLockedIncorrect calls an annotated function incorrectly (lock not held).
func (pr *ProtectedResource) CallReadDataRLockedIncorrect() int {
	return pr.readDataRLocked() // Error: Lock 'pr.rwMu' is not held before calling function requiring it.
}

// --- Atomics ---

// IncrementAtomicCorrect uses atomic operations on an atomic-only field.
func (pr *ProtectedResource) IncrementAtomicCorrect() {
	atomic.AddInt32(&pr.atomicValue, 1) // Correct: Atomic operation on atomic field.
}

// ReadAtomicCorrect uses atomic operations on an atomic-only field.
func (pr *ProtectedResource) ReadAtomicCorrect() int32 {
	return atomic.LoadInt32(&pr.atomicValue) // Correct: Atomic operation on atomic field.
}

// IncorrectDirectReadAtomic reads an atomic field directly.
func (pr *ProtectedResource) IncorrectDirectReadAtomic() int32 {
	return pr.atomicValue // Error: Field 'atomicValue' requires atomic access for reads.
}

// IncorrectDirectWriteAtomic writes an atomic field directly.
func (pr *ProtectedResource) IncorrectDirectWriteAtomic() {
	pr.atomicValue = 99 // Error: Field 'atomicValue' requires atomic access for writes.
}

// --- Mixed Atomics and Locks ---

// ReadMixedCorrectAtomic reads a mixed field atomically (allowed for reads).
func (pr *ProtectedResource) ReadMixedCorrectAtomic() int32 {
	return atomic.LoadInt32(&pr.mixedValue) // Correct: Atomic read is allowed.
}

// ReadMixedCorrectLock reads a mixed field with the lock held (allowed for reads).
func (pr *ProtectedResource) ReadMixedCorrectLock() int32 {
	pr.mu.Lock()
	v := pr.mixedValue // Correct: Lock is held for read.
	pr.mu.Unlock()
	return v
}

// WriteMixedCorrect writes a mixed field atomically with the lock held (required for writes).
func (pr *ProtectedResource) WriteMixedCorrect(v int32) {
	pr.mu.Lock()
	atomic.StoreInt32(&pr.mixedValue, v) // Correct: Lock is held and write is atomic.
	pr.mu.Unlock()
}

// WriteMixedIncorrectAtomicOnly writes a mixed field atomically *without* the lock.
func (pr *ProtectedResource) WriteMixedIncorrectAtomicOnly(v int32) {
	// Error: Lock 'pr.mu' is not held for write to mixed field 'mixedValue'.
	atomic.StoreInt32(&pr.mixedValue, v)
}

// WriteMixedIncorrectLockOnly writes a mixed field *directly* while holding the lock.
func (pr *ProtectedResource) WriteMixedIncorrectLockOnly(v int32) {
	pr.mu.Lock()
	// Error: Field 'mixedValue' requires atomic access for writes.
	pr.mixedValue = v
	pr.mu.Unlock()
}

// WriteMixedIncorrectNeither writes a mixed field *directly* and *without* the lock.
func (pr *ProtectedResource) WriteMixedIncorrectNeither(v int32) {
	// Error: Lock 'pr.mu' is not held for write to mixed field 'mixedValue'.
	// Error: Field 'mixedValue' requires atomic access for writes.
	pr.mixedValue = v
}

// --- Acquire/Release ---

// AcquireAndSet acquires the lock and sets the value.
// +checklocksacquire:pr.acquireReleaseMu
func (pr *ProtectedResource) AcquireAndSet(v int) {
	// Annotation requires lock NOT be held on entry.
	pr.acquireReleaseMu.Lock() // Acquires the lock.
	pr.acquireReleaseValue = v
	// Annotation implies lock IS held on exit.
}

// GetAndRelease reads the value and releases the lock.
// +checklocksrelease:pr.acquireReleaseMu
func (pr *ProtectedResource) GetAndRelease() int {
	// Annotation requires lock BE held on entry.
	v := pr.acquireReleaseValue
	pr.acquireReleaseMu.Unlock() // Releases the lock.
	// Annotation implies lock IS NOT held on exit.
	return v
}

// CallAcquireReleaseCorrect demonstrates the correct acquire/release cycle.
func (pr *ProtectedResource) CallAcquireReleaseCorrect() int {
	// Lock is not held here.
	pr.AcquireAndSet(1) // Acquires lock.
	// Lock is now held.
	v := pr.GetAndRelease() // Releases lock.
	// Lock is not held here.
	return v
}

// CallAcquireReleaseIncorrectAcquire calls acquire when lock is already held.
func (pr *ProtectedResource) CallAcquireReleaseIncorrectAcquire() {
	pr.acquireReleaseMu.Lock() // Manually acquire lock.
	// Error: Lock 'pr.acquireReleaseMu' is already held before calling function requiring it be acquired.
	pr.AcquireAndSet(2)
	pr.acquireReleaseMu.Unlock() // Need to unlock eventually.
}

// CallAcquireReleaseIncorrectRelease calls release when lock is not held.
func (pr *ProtectedResource) CallAcquireReleaseIncorrectRelease() int {
	// Error: Lock 'pr.acquireReleaseMu' is not held before calling function requiring it be released.
	return pr.GetAndRelease()
}

// --- Ignore/Force ---

// FunctionToIgnore contains a violation but the analyzer is told to ignore it.
// +checklocksignore
func (pr *ProtectedResource) FunctionToIgnore() {
	pr.value = -1 // Violation: Access to pr.value without holding pr.mu.
	// This violation should NOT be reported by the linter due to +checklocksignore.
}

// CallIgnoredFunction simply calls the ignored function.
func (pr *ProtectedResource) CallIgnoredFunction() {
	pr.FunctionToIgnore()
}

// ForceExample demonstrates forcing a lock state.
func (pr *ProtectedResource) ForceExample() {
	// This access is incorrect, lock mu is not held.
	pr.value = -2 // Error: Lock 'pr.mu' is not held.

	// Assume pr.mu is now held for subsequent analysis on this line.
	// Note: This doesn't actually acquire the lock!
	_ = pr.value // +checklocksforce: pr.mu

	// If the force worked, this subsequent access should NOT be reported as an error.
	pr.description = "forced"
}
