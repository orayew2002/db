# WAL Engine — Version Evolution

This document describes the evolution of the Write-Ahead Log (WAL) engine across multiple versions, showing performance improvements and architectural changes.

---

# v1 — Direct Disk Writes

In the first version, WAL logs were written directly to disk on every insert operation.

### Characteristics:

* No buffering
* No batching
* Direct file writes per operation

### Result:

* Very high syscall overhead
* Poor throughput
* Simple implementation

### Performance:

```
v1 → ~100 ops/sec
```

---

# v2 — Buffered WAL + Batch Flush

In this version, WAL was improved by introducing in-memory buffering and periodic flushing.

Logs are first stored in RAM and flushed to disk every few milliseconds/seconds.

### Improvements:

* Buffered writes (bufio.Writer)
* Batch flushing strategy
* Reduced disk I/O pressure

### Optimization Insight:

During profiling, JSON serialization (`json.Marshal/unmarshal`) was identified as a major bottleneck due to:

* High CPU usage
* Memory allocations
* GC pressure

This led to replacing JSON with a more efficient encoding strategy.

### Performance:

```
v2 → ~200K ops/sec
```

---

# v3 — Protobuf-based WAL

In the final version, JSON serialization was fully replaced with Protocol Buffers (protobuf).

This significantly reduced CPU overhead and improved serialization efficiency.

### Improvements:

* Binary serialization (protobuf)
* Reduced allocations
* Faster encode/decode pipeline
* Better memory efficiency

### Performance:

```
v3 → ~270K ops/sec
```

---

# Summary

| Version | Architecture        | Performance   |
| ------- | ------------------- | ------------- |
| v1      | Direct disk writes  | ~100 ops/sec  |
| v2      | Buffered + batching | ~200K ops/sec |
| v3      | Protobuf WAL        | ~270K ops/sec |

---

# Key Takeaways

* Disk I/O per request is extremely expensive (v1)
* Batching + buffering gives massive improvement (v2)
* Serialization format matters (JSON → protobuf) (v3)
* Current bottleneck is now memory allocations and locking, not disk I/O

---

# Future Improvements (Ideas)

* Lock-free WAL writer (ring buffer)
* Group commit (PostgreSQL-style)
* Zero-allocation encoding path
* Segment-based WAL files
* Custom binary format (faster than protobuf)
