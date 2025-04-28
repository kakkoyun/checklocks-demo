# Go `checklocks` Analyzer Demo

This repository demonstrates the usage and explores the behavior of the `checklocks` static analysis tool from the [gVisor project](https://github.com/google/gvisor). This tool helps enforce lock usage and atomic access patterns for struct fields in Go based on annotations.

**Reference:** [https://pkg.go.dev/gvisor.dev/gvisor/tools/checklocks](https://pkg.go.dev/gvisor.dev/gvisor/tools/checklocks)

## Goal

To provide a clear, runnable example showcasing:

1. How to annotate struct fields (`// +checklocks:mutexName`, `// +checkatomic`) and functions (`// +checklocks:param.mutexName`, `// +checklocksread:param.rwMutexName`, `// +checklocksacquire:...`, `// +checklocksrelease:...`) to define locking and atomic requirements.
2. How the analyzer identifies correct lock usage (including RWMutex).
3. How the analyzer enforces atomic access (`sync/atomic` required).
4. How mixed atomic and lock requirements work.
5. How acquire/release annotations (`+checklocksacquire`/`+checklocksrelease`) track lock state changes across function calls.
6. How ignore/force annotations (`+checklocksignore`/`+checklocksforce`) modify the analysis.
7. How the analyzer identifies incorrect usage (missing locks, missing atomic ops, incorrect acquire/release).
8. The behavior and limitations of the analyzer, particularly regarding scope and the `+checklocksfail` annotation.
9. How to complement `checklocks` with runtime assertions using `go-mutexasserts` for ignored functions.

## Findings & Key Behaviors Observed (Final)

* **Annotation Driven:** The analyzer strictly follows all annotations (`+checklocks:`, `+checklocksread:`, `+checkatomic`, `+checklocksacquire`, `+checklocksrelease`, `+checklocksignore`, `+checklocksforce`).
* **Locking:** Enforces `Lock`/`Unlock` and `RLock`/`RUnlock`. Checks function preconditions like `+checklocks:p.mu` and `+checklocksread:p.rwMu` at call sites.
* **Atomics:** Enforces use of `sync/atomic`. Reports direct reads/writes.
* **Mixed Mode (`+checkatomic`, `+checklocks:mu`):** Reads need lock *or* atomic. Writes need lock *and* atomic.
* **Acquire/Release:** Tracks lock state changes. `+checklocksacquire` requires the lock *not* be held on entry and assumes it *is* held on exit. `+checklocksrelease` requires the lock *be* held on entry and assumes it *is not* held on exit. The analyzer flags violations of these preconditions at call sites.
* **Ignore/Force:**
  * `+checklocksignore` on a function prevents the analyzer from checking *any* lock/atomic access rules *within that function*. This is useful when a function has internal accesses that would normally violate the rules, but you guarantee (by convention) that the function is always called under the correct lock/conditions (see `helperCalledUnderLock` example). Use with caution, as it removes safety checks for that function's body.
  * `+checklocksforce: lock` tells the analyzer to assume `lock` is held from that point onwards; it suppresses subsequent errors but can lead to warnings if the function exits with the lock seemingly held (as shown by the "return with unexpected locks held" warning). Also use with caution.
* **Scope/Call Site Analysis:** Still primarily checks call site preconditions only if the called function is annotated. It does not deeply analyze unannotated functions when checking callers.
* **`+checklocksfail` Annotation:** Confirmed useful only for asserting a violation *is* found on a specific line (e.g., calling an annotated function incorrectly), satisfying the annotation. Not effective for call sites of functions with internal-only violations or for acquire/release precondition violations.
* **Generics Support (Partial):**
  * **Real Violations:** The analyzer *does* correctly detect actual lock violations (`+checklocks`, `+checklocksread`, `+checkatomic`, etc.) in code using generics.
  * **False Positives:** It produces spurious warnings like `may require checklocks annotation for mu, used with lock held 100% of the time` for the mutex fields themselves within generic types. This appears to be a false positive specific to generics, as it doesn't occur for equivalent non-generic code.
  * **Test Annotation (`+checklocksfail`):** The `+checklocksfail` annotation used in tests does *not* seem to correctly identify expected violations when used with generic code, leading to test failures (e.g., `got 0 failures, want 1 failures`).
  * **Upstream Issue:** An issue has been opened to track this: [https://github.com/google/gvisor/issues/11671](https://github.com/google/gvisor/issues/11671)
* **`go vet` Exit Code:** Fails if any violations are found.
* **Runtime vs. Static:** Static analysis (like `checklocks`) is powerful for finding lock misuse based on annotations but cannot find all concurrency issues. Deadlocks or panics resulting from misuse (like incorrect acquire/release patterns) require runtime detection (e.g., using `-race` and `-timeout` during testing, or observing the panic). We skip tests known to deadlock or panic in this demo to allow the suite to complete.
* **Runtime Assertions (Debug Builds):** The `github.com/trailofbits/go-mutexasserts` library is used to add runtime lock assertions (`mutexasserts.AssertMutexLocked`) inside functions where static analysis is bypassed (e.g., via `+checklocksignore`). These assertions check lock state dynamically but are only active when the code is built with the `debug` tag (`go build -tags debug`, `go test -tags debug`). This provides an extra layer of safety during development/testing for assumptions made when ignoring the static checker.

This demo provides a comprehensive overview of the `checklocks` analyzer's capabilities and limitations, along with a strategy for adding runtime checks.

## Annotated Linter Output

The following shows the expected output when running `make lint`. The violations reported are intentional demonstrations of the analyzer catching incorrect patterns described above, including some false positives related to generics. The exit code is non-zero, as expected for a linter finding issues.

```text
# github.com/kakkoyun/checklocks-demo/pkg/resource
# [github.com/kakkoyun/checklocks-demo/pkg/resource]

# --- Basic Lock Violations ---
pkg/resource/resource.go:58:5: invalid field access, mu (&({param:pr}.mu)) must be locked when accessing value (locks: no locks held)
#   [Reason: Accessing `value` (`+checklocks:mu`) inside IncorrectSetData without holding `mu`.]
pkg/resource/resource.go:59:5: invalid field access, mu (&({param:pr}.mu)) must be locked when accessing description (locks: no locks held)
#   [Reason: Accessing `description` (`+checklocks:mu`) inside IncorrectSetData without holding `mu`.]
pkg/resource/resource.go:89:18: must hold pr.mu exclusively (&({param:pr}.mu)) to call setDataLocked, but not held (locks: no locks held)
#   [Reason: Calling `setDataLocked` (requires `+checklocks:pr.mu`) from IncorrectSetDataWithHelper without holding `mu`.]

# --- RWMutex / Read Lock Violations ---
pkg/resource/resource.go:114:12: invalid field access, rwMu (&({param:pr}.rwMu)) must be locked when accessing readGuardedValue (locks: no locks held)
#   [Reason: Accessing `readGuardedValue` (`+checklocks:rwMu`) inside GetReadGuardedValueIncorrect without holding `rwMu`.]
pkg/resource/resource.go:133:27: must hold pr.rwMu non-exclusively (&({param:pr}.rwMu)) to call readDataRLocked, but not held (locks: no locks held)
#   [Reason: Calling `readDataRLocked` (requires `+checklocksread:pr.rwMu`) from CallReadDataRLockedIncorrect without holding `rwMu`.]

# --- Atomic Violations ---
pkg/resource/resource.go:150:12: illegal use of atomic-only field by *ssa.UnOp instruction
#   [Reason: Reading `atomicValue` (`+checkatomic`) directly (non-atomically) in IncorrectDirectReadAtomic.]
pkg/resource/resource.go:155:5: illegal use of atomic-only field by *ssa.Store instruction
pkg/resource/resource.go:155:5: non-atomic write of field atomicValue, writes must still be atomic with locks held (locks: no locks held)
#   [Reason: Writing `atomicValue` (`+checkatomic`) directly (non-atomically) in IncorrectDirectWriteAtomic.]

# --- Mixed Mode Violations ---
pkg/resource/resource.go:183:19: unexpected call to atomic write function, is a lock missing?
#   [Reason: Writing `mixedValue` (`+checkatomic`, `+checklocks:mu`) atomically in WriteMixedIncorrectAtomicOnly *without* holding `mu`.]
pkg/resource/resource.go:190:5: illegal use of atomic-only field by *ssa.Store instruction
pkg/resource/resource.go:190:5: non-atomic write of field mixedValue, writes must still be atomic with locks held (locks: &({param:pr}.mu) exclusively)
#   [Reason: Writing `mixedValue` (`+checkatomic`, `+checklocks:mu`) directly (non-atomically) in WriteMixedIncorrectLockOnly, even though `mu` is held.]
pkg/resource/resource.go:198:5: illegal use of atomic-only field by *ssa.Store instruction
pkg/resource/resource.go:198:5: non-atomic write of field mixedValue, writes must still be atomic with locks held (locks: no locks held)
#   [Reason: Writing `mixedValue` (`+checkatomic`, `+checklocks:mu`) directly (non-atomically) *and* without holding `mu` in WriteMixedIncorrectNeither.]

# --- Acquire/Release Violations ---
pkg/resource/resource.go:236:18: attempt to acquire pr.acquireReleaseMu (&({param:pr}.acquireReleaseMu)), but already held (locks: &({param:pr}.acquireReleaseMu) exclusively)
#   [Reason: Calling `AcquireAndSet` (requires `+checklocksacquire:pr.acquireReleaseMu`) from CallAcquireReleaseIncorrectAcquire when `acquireReleaseMu` is already held.]
pkg/resource/resource.go:243:25: must hold pr.acquireReleaseMu exclusively (&({param:pr}.acquireReleaseMu)) to call GetAndRelease, but not held (locks: no locks held)
pkg/resource/resource.go:243:25: attempt to release pr.acquireReleaseMu (&({param:pr}.acquireReleaseMu)), but not held (locks: no locks held)
#   [Reason: Calling `GetAndRelease` (requires `+checklocksrelease:pr.acquireReleaseMu`) from CallAcquireReleaseIncorrectRelease when `acquireReleaseMu` is not held.]

# --- Force Example Violation ---
pkg/resource/resource.go:263:5: invalid field access, mu (&({param:pr}.mu)) must be locked when accessing value (locks: no locks held)
#   [Reason: Accessing `value` (`+checklocks:mu`) in ForceExample before the `+checklocksforce` annotation.]

# --- Force Example Side Effect ---
-: return with unexpected locks held (locks: &({param:pr}.mu) exclusively)
#   [Reason: The `+checklocksforce: pr.mu` in ForceExample told the analyzer `mu` was held, but it was never released, so the analyzer thinks the function returns holding the lock.]

# --- Generic Resource Violations (pkg/genericresource) ---
# github.com/kakkoyun/checklocks-demo/pkg/genericresource
# [github.com/kakkoyun/checklocks-demo/pkg/genericresource]
pkg/genericresource/generic.go:15:2: may require checklocks annotation for mu, used with lock held 100% of the time
#   [Reason: False positive warning specific to generics. See https://github.com/google/gvisor/issues/11671]
pkg/genericresource/generic.go:23:2: may require checklocks annotation for rwMu, used with lock held 100% of the time
#   [Reason: False positive warning specific to generics. See https://github.com/google/gvisor/issues/11671]
pkg/genericresource/generic.go:34:2: may require checklocks annotation for acquireReleaseMu, used with lock held 100% of the time
#   [Reason: False positive warning specific to generics. See https://github.com/google/gvisor/issues/11671]
# (Note: The analyzer correctly identifies real violations in generic code, similar to the non-generic examples above, but those are omitted here for brevity as they are covered by the original pkg/resource examples.)

# --- Note: Ignored Violations ---
# - No error reported for access within FunctionToIgnore due to `+checklocksignore`.
# - No error reported for access after `+checklocksforce` in ForceExample.
# - No error reported for call to setDataLocked in TestDirectCallToSetDataLocked due to `+checklocksfail`.

```

## Usage

1. **Install dependencies:**

    ```bash
    go mod tidy
    ```

2. **Install `checklocks` vet tool:**

    ```bash
    make install-vettool
    ```

3. **Run linter (finds expected violations):**

    ```bash
    # Note: The -tags debug flag enables runtime assertions (if any) during vet checks
    # that might involve running code, though checklocks is static.
    make lint
    ```

4. **Run tests (with race detector and debug assertions enabled):**

    ```bash
    # Note: The -tags debug flag enables the go-mutexasserts runtime checks.
    make test
    ```

5. **Clean:**

    ```bash
    make clean
    ```

## Exploring the Code

* `pkg/resource/resource.go`: Contains the `ProtectedResource` struct with various annotations and methods demonstrating correct/incorrect usage.
* `pkg/resource/resource_test.go`: Contains test cases, including some using `+checklocksfail` to assert expected linter violations and others verifying `go-mutexasserts` behavior with the `debug` tag.
* `pkg/genericresource/generic.go`: Contains a generic version (`GenericResource[T]`) used to test the analyzer's behavior with generics.
* `pkg/genericresource/generic_test.go`: Contains basic tests for the generic resource.
* `Makefile`: Defines targets for installation, linting, testing, and cleaning.
