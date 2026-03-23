# Xxlang Concurrency Programming

Xxlang provides Go-style concurrency primitives for building concurrent applications.

## Overview

Xxlang's concurrency model is inspired by Go's CSP (Communicating Sequential Processes) model:

- **Goroutines** - Lightweight concurrent execution with `run` keyword
- **Tubes** - Typed channels for communication between goroutines
- **Select** - Multiplex operations on multiple tubes
- **Context** - Timeout and cancellation management
- **Sync Primitives** - Mutex, RWMutex, WaitGroup, Once, Cond, AtomicInt

## Goroutines

Use the `run` keyword to start a new goroutine:

```xxl
// Run anonymous block
run {
    sleep(100)
    pln("Background task completed")
}

// Run function with arguments
run worker(1, 2, 3)
```

### Example: Parallel Execution

```xxl
var counter = newAtomic(0)

// Start multiple goroutines
for (var i = 0; i < 10; i = i + 1) {
    run {
        counter.add(1)
    }
}

sleep(50)
pln("Counter:", counter.load())  // Output: Counter: 10
```

### Notes

- The `run` statement returns `null`
- Goroutines run concurrently and independently
- Use sync primitives (WaitGroup, Tube) to coordinate

## Tubes (Channels)

Tubes are typed channels for communication between goroutines.

### Creating Tubes

```xxl
// Untyped tube with buffer
var tube = makeTube(10)

// Typed tube (elements must match type)
var intTube = makeTube("INT", 5)
var strTube = makeTube("STRING", 3)

// Unbuffered tube (synchronous)
var syncTube = makeTube(0)
```

### Supported Types for Typed Tubes

| Type String | Description |
|-------------|-------------|
| `"INT"` | Integer values |
| `"FLOAT"` | Floating-point values |
| `"STRING"` | String values |
| `"BOOL"` | Boolean values |
| `"ARRAY"` | Array values |
| `"MAP"` | Map values |

### Sending and Receiving

Use the `<-` operator for tube operations:

```xxl
var tube = makeTube(1)

// Send (blocking)
tube <- 42

// Receive (blocking)
var val = <- tube
pln(val)  // Output: 42
```

### Buffered vs Unbuffered

```xxl
// Buffered tube - can hold multiple values
var buffered = makeTube(3)
buffered <- 1
buffered <- 2
buffered <- 3  // All sends succeed immediately

// Unbuffered tube - sender blocks until receiver
var unbuffered = makeTube(0)
run {
    unbuffered <- "hello"  // Blocks until received
}
var msg = <- unbuffered
```

### Non-blocking Operations

```xxl
var tube = makeTube(1)

// TrySend - non-blocking send, returns boolean
var sent = tubeTrySend(tube, 42)
if (sent) {
    pln("Sent successfully")
}

// TryReceive - non-blocking receive
var result = tubeTryRecv(tube)
// Returns [value, received, open]
var value = result[0]
var received = result[1]
var open = result[2]

// Check tube state
pln("Length:", tubeLen(tube))
pln("Capacity:", tubeCap(tube))
pln("Closed:", tubeClosed(tube))
```

### Closing Tubes

```xxl
var tube = makeTube(2)
tube <- 1
tube <- 2
closeTube(tube)

// Can still receive buffered values
var a = <- tube  // 1
var b = <- tube  // 2
var c = <- tube  // null (tube closed and empty)
```

### Tube Methods

| Method | Description |
|--------|-------------|
| `tube.send(value)` | Send value (blocking) |
| `tube.receive()` | Receive value, returns `[value, ok]` |
| `tube.trySend(value)` | Non-blocking send, returns `[sent, ok]` |
| `tube.tryReceive()` | Non-blocking receive, returns `[value, received, open]` |
| `tube.close()` | Close the tube |
| `tube.len()` | Number of buffered elements |
| `tube.cap()` | Buffer capacity |
| `tube.isClosed()` | Check if tube is closed |

## Select Statement

Select lets you wait on multiple tube operations simultaneously.

### Basic Select

```xxl
var tube1 = makeTube(1)
var tube2 = makeTube(1)

select {
case v = <- tube1:
    pln("From tube1:", v)
case v = <- tube2:
    pln("From tube2:", v)
}
```

### Select with Default (Non-blocking)

```xxl
var tube = makeTube(1)

select {
case v = <- tube:
    pln("Received:", v)
default:
    pln("No data available")
}
```

### Select with Send

```xxl
var tube = makeTube(1)

select {
case tube <- 42:
    pln("Sent 42")
default:
    pln("Tube full, cannot send")
}
```

### Multiple Cases with Default

```xxl
var tube1 = makeTube(1)
var tube2 = makeTube(1)

select {
case v = <- tube1:
    pln("tube1:", v)
case v = <- tube2:
    pln("tube2:", v)
default:
    pln("No data")
}
```

## Context

Context provides timeout and cancellation for goroutines.

### Context with Timeout

```xxl
var ctx = contextWithTimeout(null, 1000)  // 1 second timeout

run {
    select {
    case <- ctx.done():
        pln("Context timed out")
    }
}

sleep(1100)
pln("Is done:", ctx.isDone())  // true
```

### Context with Cancel

```xxl
var ctx = contextWithCancel(null)

run {
    select {
    case <- ctx.done():
        pln("Cancelled!")
    }
}

sleep(100)
ctx.cancel()  // Signal cancellation
```

### Context with Deadline

```xxl
var deadline = now(1) + 5000  // 5 seconds from now
var ctx = contextWithDeadline(null, deadline)

select {
case <- ctx.done():
    pln("Deadline reached:", ctx.err())
}
```

### Context in Select

```xxl
var tube = makeTube(1)
var ctx = contextWithTimeout(null, 500)

run {
    sleep(1000)
    tube <- "delayed message"
}

select {
case v = <- tube:
    pln("Received:", v)
case <- ctx.done():
    pln("Timeout!")
}
```

### Context Methods

| Method | Description |
|--------|-------------|
| `ctx.done()` | Returns tube that closes when context is done |
| `ctx.cancel()` | Cancel the context |
| `ctx.err()` | Returns error string if done |
| `ctx.isDone()` | Returns true if context is done |
| `ctx.deadline()` | Returns deadline timestamp (ms) or null |
| `ctx.deadlineStr()` | Returns formatted deadline string |

## Sync Primitives

### Mutex

```xxl
var mu = newMutex()
var counter = 0

run {
    mu.lock()
    counter = counter + 1
    mu.unlock()
}

run {
    mu.lock()
    counter = counter + 1
    mu.unlock()
}

sleep(100)
pln("Counter:", counter)  // 2
```

#### Mutex Methods

| Method | Description |
|--------|-------------|
| `mu.lock()` | Acquire lock (blocking) |
| `mu.unlock()` | Release lock |
| `mu.tryLock()` | Try to acquire lock (non-blocking), returns boolean |

### RWMutex

```xxl
var rwmu = newRWMutex()
var data = {"name": "test"}

// Multiple readers can hold RLock simultaneously
rwmu.rLock()
pln(data.name)
rwmu.rUnlock()

// Single writer - exclusive lock
rwmu.lock()
data.value = 42
rwmu.unlock()
```

#### RWMutex Methods

| Method | Description |
|--------|-------------|
| `rwmu.lock()` | Acquire write lock (exclusive) |
| `rwmu.unlock()` | Release write lock |
| `rwmu.tryLock()` | Try to acquire write lock (non-blocking) |
| `rwmu.rLock()` | Acquire read lock (shared) |
| `rwmu.rUnlock()` | Release read lock |
| `rwmu.tryRLock()` | Try to acquire read lock (non-blocking) |

### WaitGroup

```xxl
var wg = newWaitGroup()
var results = []

for (var i = 0; i < 5; i = i + 1) {
    wg.add(1)
    run {
        results = push(results, i * 2)
        wg.done()
    }
}

wg.wait()  // Wait for all goroutines
pln("Results:", results)
```

#### WaitGroup Methods

| Method | Description |
|--------|-------------|
| `wg.add(delta)` | Add delta to counter |
| `wg.done()` | Decrement counter by 1 |
| `wg.wait()` | Block until counter is 0 |

### Once

```xxl
var once = newOnce()
var initialized = false

func initFunc() {
    initialized = true
    pln("Initializing...")
}

// Will only execute once
once.do(initFunc)
once.do(initFunc)  // No effect
```

#### Once Methods

| Method | Description |
|--------|-------------|
| `once.do(func)` | Execute function only once |

### Cond

```xxl
var mu = newMutex()
var cond = newCond(mu)
var ready = false
var data = null

// Waiter
run {
    mu.lock()
    while (!ready) {
        cond.wait()
    }
    pln("Got data:", data)
    mu.unlock()
}

// Signaler
sleep(100)
mu.lock()
data = "hello"
ready = true
cond.signal()
mu.unlock()
```

#### Cond Methods

| Method | Description |
|--------|-------------|
| `cond.wait()` | Wait for signal (must hold lock) |
| `cond.signal()` | Wake one waiter |
| `cond.broadcast()` | Wake all waiters |
| `cond.lock()` | Acquire associated mutex |
| `cond.unlock()` | Release associated mutex |

### AtomicInt

```xxl
var counter = newAtomic(0)

// Atomic add returns new value
var v1 = counter.add(5)  // 5
var v2 = counter.add(3)  // 8

// Load current value
var current = counter.load()  // 8

// Store new value
counter.store(100)

// Swap returns old value
var old = counter.swap(200)  // old = 100, counter = 200

// Compare and swap
var swapped = counter.compareAndSwap(200, 300)  // true, counter = 300
var notSwapped = counter.compareAndSwap(100, 400)  // false, counter = 300
```

#### AtomicInt Methods

| Method | Description |
|--------|-------------|
| `atomic.add(delta)` | Add delta atomically, returns new value |
| `atomic.load()` | Load current value |
| `atomic.store(value)` | Store new value |
| `atomic.swap(new)` | Swap with new value, returns old value |
| `atomic.compareAndSwap(old, new)` | CAS operation, returns boolean |

## Built-in Functions

### Tube Functions

| Function | Description |
|----------|-------------|
| `makeTube([type], buffer)` | Create a new tube |
| `closeTube(tube)` | Close a tube |
| `tubeLen(tube)` | Get number of buffered elements |
| `tubeCap(tube)` | Get buffer capacity |
| `tubeClosed(tube)` | Check if tube is closed |
| `tubeSend(tube, value)` | Send value to tube |
| `tubeRecv(tube)` | Receive from tube, returns `[value, ok]` |
| `tubeTrySend(tube, value)` | Non-blocking send, returns `[sent, ok]` |
| `tubeTryRecv(tube)` | Non-blocking receive, returns `[value, received, open]` |

### Context Functions

| Function | Description |
|----------|-------------|
| `newContext()` | Create background context |
| `contextWithTimeout(parent, ms)` | Context with timeout in milliseconds |
| `contextWithCancel(parent)` | Context with cancel capability |
| `contextWithDeadline(parent, timestamp)` | Context with deadline (ms timestamp) |
| `contextCancel(ctx)` | Cancel context |
| `contextDone(ctx)` | Get done tube |
| `contextErr(ctx)` | Get error string |
| `contextIsDone(ctx)` | Check if done |
| `contextDeadline(ctx)` | Get deadline timestamp |

### Sync Functions

| Function | Description |
|----------|-------------|
| `newMutex()` | Create new mutex |
| `newRWMutex()` | Create new RWMutex |
| `newWaitGroup()` | Create new WaitGroup |
| `newOnce()` | Create new Once |
| `newCond([mutex])` | Create new Cond (optional mutex) |
| `newAtomic([initial])` | Create new atomic int with optional initial value |

## Standard Library: concurrent Module

Import the `concurrent` module for additional concurrency utilities:

```xxl
import "concurrent"

// Module provides same functions as built-ins
var tube = concurrent.makeTube(10)
var mu = concurrent.newMutex()
```

## Complete Example: Worker Pool

```xxl
// Worker pool with tubes and WaitGroup
var jobs = makeTube(10)
var results = makeTube(10)
var wg = newWaitGroup()

// Worker function
func worker(id) {
    defer wg.done()

    while (true) {
        var job = tubeTryRecv(jobs)
        if (!job[1]) {
            break  // No more jobs
        }
        var n = job[0]
        results <- n * n
    }
}

// Start 3 workers
for (var i = 0; i < 3; i = i + 1) {
    wg.add(1)
    run worker(i)
}

// Send 5 jobs
for (var i = 1; i <= 5; i = i + 1) {
    jobs <- i
}
closeTube(jobs)

// Wait for workers to finish
wg.wait()
closeTube(results)

// Collect results
var total = 0
while (true) {
    var r = tubeTryRecv(results)
    if (!r[1]) {
        break
    }
    pln("Result:", r[0])
    total = total + r[0]
}

pln("Total:", total)  // 1+4+9+16+25 = 55
```

## Complete Example: Producer-Consumer

```xxl
var items = makeTube(5)
var done = makeTube(1)

// Producer
run {
    for (var i = 1; i <= 10; i = i + 1) {
        items <- i
        pln("Produced:", i)
    }
    closeTube(items)
}

// Consumer
run {
    while (true) {
        var item = <- items
        if (item == null) {
            break
        }
        pln("Consumed:", item)
    }
    done <- 1
}

// Wait for consumer
<- done
pln("Done!")
```

## Complete Example: Timeout Pattern

```xxl
var tube = makeTube(1)
var ctx = contextWithTimeout(null, 500)

// Simulate slow operation
run {
    sleep(1000)
    tube <- "result"
}

select {
case v = <- tube:
    pln("Got result:", v)
case <- ctx.done():
    pln("Operation timed out!")
}
```

## Best Practices

1. **Always close tubes** when producers are done to prevent goroutine leaks
2. **Use contexts** for timeout and cancellation instead of manual timeouts
3. **Use select with default** for non-blocking operations
4. **Prefer buffered tubes** for better throughput when order doesn't matter
5. **Use WaitGroups** to wait for multiple goroutines to complete
6. **Use atomic operations** for simple counters instead of mutex
7. **Use RWMutex** when reads are more frequent than writes
8. **Avoid sharing memory** - prefer communication via tubes
