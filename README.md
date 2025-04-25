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

## Findings & Key Behaviors Observed (Final)

* **Annotation Driven:** The analyzer strictly follows all annotations (`+checklocks:`, `+checklocksread:`, `+checkatomic`, `+checklocksacquire`, `+checklocksrelease`, `+checklocksignore`, `+checklocksforce`).
* **Locking:** Enforces `Lock`/`Unlock` and `RLock`/`RUnlock`. Checks function preconditions like `+checklocks:p.mu` and `+checklocksread:p.rwMu` at call sites.
* **Atomics:** Enforces use of `sync/atomic`. Reports direct reads/writes.
* **Mixed Mode (`+checkatomic`, `+checklocks:mu`):** Reads need lock *or* atomic. Writes need lock *and* atomic.
* **Acquire/Release:** Tracks lock state changes. `+checklocksacquire` requires the lock *not* be held on entry and assumes it *is* held on exit. `+checklocksrelease` requires the lock *be* held on entry and assumes it *is not* held on exit. The analyzer flags violations of these preconditions at call sites.
* **Ignore/Force:** `+checklocksignore` successfully suppresses violation reports within the annotated function. `+checklocksforce: lock` tells the analyzer to assume `lock` is held from that point onwards; it suppresses subsequent errors but can lead to warnings if the function exits with the lock seemingly held (as shown by the "return with unexpected locks held" warning).
* **Scope/Call Site Analysis:** Still primarily checks call site preconditions only if the called function is annotated. It does not deeply analyze unannotated functions when checking callers.
* **`+checklocksfail` Annotation:** Confirmed useful only for asserting a violation *is* found on a specific line (e.g., calling an annotated function incorrectly), satisfying the annotation. Not effective for call sites of functions with internal-only violations or for acquire/release precondition violations.
* **`go vet` Exit Code:** Fails if any violations are found.
* **Runtime vs. Static:** Static analysis (like `checklocks`) is powerful for finding lock misuse based on annotations but cannot find all concurrency issues. Deadlocks or panics resulting from misuse (like incorrect acquire/release patterns) require runtime detection (e.g., using `-race` and `-timeout` during testing, or observing the panic). We skip tests known to deadlock or panic in this demo to allow the suite to complete.

This demo provides a comprehensive overview of the `checklocks` analyzer's capabilities and limitations.
