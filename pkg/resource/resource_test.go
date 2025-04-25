package resource

import (
	"testing"
)

// Helper to create a new resource for tests, updated for new fields
func newTestResource() *ProtectedResource {
	// Initial values: value=0, readGuardedValue=10, atomicValue=20, mixedValue=30, acqRelValue=40, desc="initial", id="id-0"
	return NewProtectedResource(0, 10, 40, 20, 30, "initial", "id-0")
}

// TestCorrectSetData verifies correct locking for setting data.
func TestCorrectSetData(t *testing.T) {
	pr := newTestResource()
	pr.SetData(1, "updated")
	val, desc := pr.GetData()
	if val != 1 || desc != "updated" {
		t.Errorf("SetData failed: expected 1/updated, got %d/%s", val, desc)
	}
}

// TestIncorrectSetData expects a checklocks failure for setting data without locking.
func TestIncorrectSetData(t *testing.T) {
	pr := newTestResource()
	pr.IncorrectSetData(2, "bad update") // Linter should report violation within IncorrectSetData
	// Note: The actual values might be updated due to the race, but the test focuses on the lint failure.
}

// TestCorrectSetDataWithHelper verifies correct locking when using the helper.
func TestCorrectSetDataWithHelper(t *testing.T) {
	pr := newTestResource()
	pr.SetDataWithHelper(3, "helper update")
	val, desc := pr.GetData()
	if val != 3 || desc != "helper update" {
		t.Errorf("SetDataWithHelper failed: expected 3/helper update, got %d/%s", val, desc)
	}
}

// TestIncorrectSetDataWithHelper expects a checklocks failure for calling the locked helper without locking.
func TestIncorrectSetDataWithHelper(t *testing.T) {
	pr := newTestResource()
	pr.IncorrectSetDataWithHelper(4, "bad helper update") // Linter should report violation within IncorrectSetDataWithHelper
}

// TestDirectCallToSetDataLocked expects a checklocks failure for calling the locked helper directly without the lock.
// Now we can call the unexported method directly.
func TestDirectCallToSetDataLocked(t *testing.T) {
	pr := newTestResource()
	// Directly call the unexported method requiring a lock, without holding it.
	pr.setDataLocked(5, "direct bad update") // +checklocksfail expected direct call violation on unexported annotated function setDataLocked
}

// TestIDAccess verifies that accessing the unguarded ID field works without locks and without analyzer errors.
func TestIDAccess(t *testing.T) {
	pr := newTestResource()
	pr.SetID("new-id-5")
	newID := pr.GetID()
	if newID != "new-id-5" {
		t.Errorf("SetID/GetID failed: expected new-id-5, got %s", newID)
	}
	// No +checklocksfail annotation here, as access is correct.
}

// TestRWMutexCorrectRead verifies correct reading using RWMutex.
func TestRWMutexCorrectRead(t *testing.T) {
	pr := newTestResource()
	v := pr.GetReadGuardedValueCorrect()
	if v != 10 { // Initial value from newTestResource
		t.Errorf("GetReadGuardedValueCorrect failed: expected 10, got %d", v)
	}
}

// TestRWMutexIncorrectRead expects a checklocks failure for reading using RWMutex without locking.
func TestRWMutexIncorrectRead(t *testing.T) {
	pr := newTestResource()
	_ = pr.GetReadGuardedValueIncorrect() // Linter should report violation within GetReadGuardedValueIncorrect
}

// TestChecklocksReadCorrect verifies correct reading using readDataRLocked.
func TestChecklocksReadCorrect(t *testing.T) {
	pr := newTestResource()
	v := pr.CallReadDataRLockedCorrect()
	if v != 10 {
		t.Errorf("CallReadDataRLockedCorrect failed: expected 10, got %d", v)
	}
}

// TestChecklocksReadIncorrect expects a checklocks failure for reading using readDataRLocked without locking.
func TestChecklocksReadIncorrect(t *testing.T) {
	pr := newTestResource()
	_ = pr.CallReadDataRLockedIncorrect() // Linter should report violation within CallReadDataRLockedIncorrect
}

// TestAtomicCorrect verifies correct atomic operations.
func TestAtomicCorrect(t *testing.T) {
	pr := newTestResource()
	initial := pr.ReadAtomicCorrect()
	if initial != 20 { // Initial value
		t.Errorf("ReadAtomicCorrect initial failed: expected 20, got %d", initial)
	}
	pr.IncrementAtomicCorrect()
	final := pr.ReadAtomicCorrect()
	if final != 21 {
		t.Errorf("Increment/ReadAtomicCorrect final failed: expected 21, got %d", final)
	}
}

// TestAtomicIncorrectRead expects a checklocks failure for incorrect atomic read.
func TestAtomicIncorrectRead(t *testing.T) {
	pr := newTestResource()
	_ = pr.IncorrectDirectReadAtomic() // Linter should report violation within IncorrectDirectReadAtomic
}

// TestAtomicIncorrectWrite expects a checklocks failure for incorrect atomic write.
func TestAtomicIncorrectWrite(t *testing.T) {
	pr := newTestResource()
	pr.IncorrectDirectWriteAtomic() // Linter should report violation within IncorrectDirectWriteAtomic
}

// TestMixedModeCorrectRead verifies correct mixed mode access.
func TestMixedModeCorrectRead(t *testing.T) {
	pr := newTestResource()
	// Read atomically
	vAtomic := pr.ReadMixedCorrectAtomic()
	if vAtomic != 30 { // Initial value
		t.Errorf("ReadMixedCorrectAtomic failed: expected 30, got %d", vAtomic)
	}
	// Read with lock
	vLock := pr.ReadMixedCorrectLock()
	if vLock != 30 {
		t.Errorf("ReadMixedCorrectLock failed: expected 30, got %d", vLock)
	}
}

// TestMixedModeCorrectWrite verifies correct mixed mode access.
func TestMixedModeCorrectWrite(t *testing.T) {
	pr := newTestResource()
	pr.WriteMixedCorrect(31)
	v := pr.ReadMixedCorrectAtomic() // Read back to verify
	if v != 31 {
		t.Errorf("WriteMixedCorrect failed: expected 31, got %d", v)
	}
}

// TestMixedModeIncorrectWriteAtomicOnly expects a checklocks failure for incorrect mixed mode write (atomic only).
func TestMixedModeIncorrectWriteAtomicOnly(t *testing.T) {
	pr := newTestResource()
	pr.WriteMixedIncorrectAtomicOnly(32) // Linter should report violation within WriteMixedIncorrectAtomicOnly
}

// TestMixedModeIncorrectWriteLockOnly expects a checklocks failure for incorrect mixed mode write (lock only).
func TestMixedModeIncorrectWriteLockOnly(t *testing.T) {
	pr := newTestResource()
	pr.WriteMixedIncorrectLockOnly(33) // Linter should report violation within WriteMixedIncorrectLockOnly
}

// TestMixedModeIncorrectWriteNeither expects a checklocks failure for incorrect mixed mode write (neither).
func TestMixedModeIncorrectWriteNeither(t *testing.T) {
	pr := newTestResource()
	pr.WriteMixedIncorrectNeither(34) // Linter should report violation within WriteMixedIncorrectNeither
}

// --- Acquire/Release Tests (New) ---

func TestAcquireReleaseCorrect(t *testing.T) {
	pr := newTestResource()
	v := pr.CallAcquireReleaseCorrect() // Calls AcquireAndSet(1), then GetAndRelease()
	if v != 1 {                         // Should get the value set by AcquireAndSet
		t.Errorf("CallAcquireReleaseCorrect failed: expected 1, got %d", v)
	}
	// We should also be able to acquire again now
	pr.AcquireAndSet(100)
	finalVal := pr.GetAndRelease()
	if finalVal != 100 {
		t.Errorf("CallAcquireReleaseCorrect second cycle failed: expected 100, got %d", finalVal)
	}
}

// TestAcquireReleaseIncorrectAcquire is skipped because it causes a deadlock.
// The underlying faulty pattern in CallAcquireReleaseIncorrectAcquire is still
// correctly flagged by `make lint`.
func TestAcquireReleaseIncorrectAcquire(t *testing.T) {
	t.Skip("Skipping test: Known to cause deadlock due to incorrect acquire pattern.")
	pr := newTestResource()
	pr.CallAcquireReleaseIncorrectAcquire() // Linter should report violation within CallAcquireReleaseIncorrectAcquire
}

// TestAcquireReleaseIncorrectRelease is skipped because it causes a panic.
// The underlying faulty pattern in CallAcquireReleaseIncorrectRelease is still
// correctly flagged by `make lint`.
func TestAcquireReleaseIncorrectRelease(t *testing.T) {
	t.Skip("Skipping test: Known to cause panic (unlock of unlocked mutex) due to incorrect release pattern.")
	pr := newTestResource()
	_ = pr.CallAcquireReleaseIncorrectRelease() // Linter should report violation within CallAcquireReleaseIncorrectRelease
}

// --- Ignore/Force Tests (New) ---

func TestIgnore(t *testing.T) {
	pr := newTestResource()
	pr.CallIgnoredFunction() // Calls FunctionToIgnore internally
	// No +checklocksfail here. The linter should NOT report the violation
	// within FunctionToIgnore because of the +checklocksignore annotation.
	// We can check if the value was actually set (proving the code ran)
	pr.mu.Lock()
	val := pr.value // Need lock for this direct access
	pr.mu.Unlock()
	if val != -1 {
		t.Errorf("FunctionToIgnore did not seem to run: expected value -1, got %d", val)
	}
}

func TestForce(t *testing.T) {
	pr := newTestResource()
	pr.ForceExample() // Contains internal violation, then +checklocksforce
	// No +checklocksfail needed here. Linter should report the first violation
	// within ForceExample, but not the second one (after the force).
	// Verify the second write happened:
	pr.mu.Lock()
	desc := pr.description // Need lock for this direct access
	pr.mu.Unlock()
	if desc != "forced" {
		t.Errorf("ForceExample second write did not seem to happen: expected desc 'forced', got '%s'", desc)
	}
}
