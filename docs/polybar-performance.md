# Polybar Performance Optimization

## Current State

Polybar calls `openriot` for various modules. Some calls are expensive (parse i3 tree, read system files).

### Call Frequency (as of v1.5)

| Module | Flag | Interval | Calls/sec |
|--------|------|----------|-----------|
| workspaces x4 | --workspace-icons-all | 2 sec | 2/sec |
| window-title | --window-title | 2 sec | 0.5/sec |
| cpu | --polybar-metrics | 30 sec | 0.03/sec |
| memory | --polybar-memory | 30 sec | 0.03/sec |
| volume | --polybar-volume | 5 sec | 0.2/sec |
| network | --network | 30 sec | 0.03/sec |
| battery | --battery | 30 sec | 0.03/sec |
| night-light | --night-light-status | 30 sec | 0.03/sec |
| wireguard | --wireguard-status | 10 sec | 0.1/sec |
| weather | --weather | 1800 sec | ~0 |

**Total: ~3 calls/sec continuously**

## The Problem

### Expensive Operations

1. **i3 tree parsing** (workspaces, window-title)
   - Calls `i3-msg -t get_tree`
   - Parses entire JSON tree (Python or Go)
   - Runs 2.5+ times per second

2. **System metrics** (cpu, memory)
   - Reads `/dev/swhog` or sysinfo files
   - Low cost but still repeated

### Why It Matters

- Each call spawns a new process
- i3 tree parsing is O(n) on window count
- With many windows, CPU usage increases
- Battery impact on laptops

## Potential Solutions

### 1. Cached Daemon (Recommended)

Run a background daemon that:
- Listens for i3 events via IPC socket
- Updates metrics on change
- Writes to cache files (`/tmp/openriot/*.cache`)

Polybar modules just `cat` the cache files - no spawning.

```
# Example cache structure
/tmp/openriot/
├── window-title.cache    # Current focused window
├── workspaces.cache      # JSON of all workspaces with icons
├── cpu.cache            # CPU percentage
└── memory.cache        # Memory percentage
```

**Pros:** Real-time updates only when needed, minimal CPU  
**Cons:** Need to manage daemon lifecycle

### 2. Increase Intervals Further

Simple but degrades UX.

### 3. Use Polybar IPC

Polybar's `type = custom/ipc` module:
- One persistent connection
- Updates pushed from daemon
- More complex to implement

**Pros:** True real-time, efficient  
**Cons:** Requires daemon + polybar config changes

### 4. Combine Metrics

Single `--polybar-all` flag that outputs:
```
cpu: 15%
mem: 42%
window: "some title"
```

One call, multiple modules read the same output.

**Pros:** Reduces calls from 6+ to ~4  
**Cons:** Still doesn't solve i3 parsing frequency

## Implementation Priority

1. **Short term:** Combine `--polybar-metrics` into one call (cpu+memory together)
2. **Medium term:** Cache daemon for i3 tree
3. **Long term:** Full IPC solution

## Related Files

- `config/polybar/config.ini` - module definitions
- `source/polybar/metrics.go` - CPU/RAM modules
- `source/windowtitle/windowtitle.go` - window title
- `source/workspaceicons/workspaceicons.go` - workspace icons
