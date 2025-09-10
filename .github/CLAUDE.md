# CLAUDE.md - go-cache Library Guide

## 🎯 Quick Reference

**go-cache** is a Redis cache dependency system built on the redigo package, providing advanced features like cache dependencies, NewRelic integration, distributed locking, and connection pooling.

### Core Commands
```bash
magex test         # Run all tests
magex test:race    # Run tests with race detector
magex lint         # Lint code
magex format:fix   # Format code
magex build        # Build project
```

## 📁 Architecture Overview

### Core Files Structure
```
cache.go         # Main cache operations (Get, Set, Delete, etc.)
pool.go          # Connection pool management & client setup
dependency.go    # Cache dependency system (KillByDependency)
scripts.go       # Lua script registration & management
redis_lock.go    # Distributed locking implementation
hash.go          # Redis hash operations
sets.go          # Redis set operations
nrredis/         # NewRelic integration wrapper
  ├── pool.go    # Wrapped pool with NR monitoring
  ├── conn.go    # Wrapped connections with NR segments
  └── options.go # Configuration options
examples/        # Usage examples for each feature
```

## 🔧 Key Components

### 1. Client & Connection Management
- **Client**: Contains redis pool, scripts, and dependency SHA
- **Context-aware**: All operations support `context.Context`
- **Dual methods**: Regular (auto-manages connections) and `Raw` (use existing connection)

### 2. Cache Dependencies
- **Core feature**: Keys can depend on other keys
- **Pattern**: `depend:<key>` stores dependents as Redis sets
- **Auto-cleanup**: Deleting a key removes all dependent keys via Lua script

### 3. NewRelic Integration
- **Optional**: Enabled via `newRelicEnabled` parameter in `Connect()`
- **Automatic**: Wraps connections to create datastore segments
- **Zero-config**: Works transparently with existing operations

## 🚀 Common Usage Patterns

### Basic Connection Setup
```go
client, err := cache.Connect(
    ctx,
    "redis://localhost:6379",
    maxActive, idleConnections,
    maxConnLifetime, idleTimeout,
    dependencyMode, newRelicEnabled,
)
```

### Cache Operations with Dependencies
```go
// Set with dependencies
cache.Set(ctx, client, "user:123", userData, "users", "active_users")

// Delete removes key and all dependents
cache.Delete(ctx, client, "users") // Also removes "user:123"
```

### Using Raw Methods (Connection Reuse)
```go
conn, err := client.GetConnectionWithContext(ctx)
defer client.CloseConnection(conn)

cache.SetRaw(conn, "key", "value", "dependency")
cache.GetRaw(conn, "key")
```

## 🧪 Testing Guidelines

### Required Tests Before Commits
```bash
magex test         # Fast tests
magex test:race    # Race condition detection
magex lint         # Code quality
```

### Testing Patterns
- **Mock Redis**: Uses `redigomock` for unit tests
- **Real Redis**: Some tests require actual Redis instance
- **Coverage**: High test coverage expected (>90%)
- **Race detection**: Critical for concurrent operations

## ⚠️ Critical Considerations

### Connection Management
- **Always close**: Use `defer client.CloseConnection(conn)` with Raw methods
- **Context**: Prefer `GetConnectionWithContext()` over deprecated `GetConnection()`
- **Pool cleanup**: Client handles connection pool lifecycle

### Dependency System
- **Lua scripts**: Dependency deletion uses pre-registered Lua scripts
- **SHA validation**: Script SHA is verified on registration
- **Transactional**: Dependency linking uses MULTI/EXEC for atomicity

### Error Handling
- **Redis errors**: Check for `redis.ErrNil` for missing keys
- **Pool errors**: Handle `ErrRedisPoolNil` for closed pools
- **Lock errors**: `ErrLockMismatch` for distributed lock conflicts

## 🔒 Security & Performance

### Security
- **No secrets**: Never log or expose Redis connection strings
- **Authentication**: Handle via Redis URL format
- **Script injection**: Use pre-registered scripts only

### Performance
- **Connection reuse**: Use Raw methods for multiple operations
- **Dependency overhead**: Consider impact of dependency tracking
- **NewRelic cost**: Monitor overhead when enabled

## 🛠️ Development Workflow

### Making Changes
1. **Read AGENTS.md**: Follow complete development guidelines
2. **Test locally**: Run `magex test` before committing
3. **Format**: Use `magex format:fix` for consistent style
4. **Dependencies**: Run `magex tidy` after adding imports

### Adding Features
- **Follow patterns**: Use existing client/raw method pairs
- **Add tests**: Both mock and integration tests
- **Update examples**: Add usage examples if applicable
- **Document**: Update relevant docs and comments

### Common Redis Commands Map
```go
GetCommand           = "GET"        # cache.Get()
SetCommand           = "SET"        # cache.Set()
DeleteCommand        = "DEL"        # cache.Delete()
HashGetCommand       = "HGET"       # cache.HashGet()
HashKeySetCommand    = "HSET"       # cache.HashSet()
AddToSetCommand      = "SADD"       # cache.SetAdd()
MembersCommand       = "SMEMBERS"   # cache.SetMembers()
```

## 🚨 Watch Out For

- **Dependency cycles**: Avoid circular dependencies between keys
- **Script changes**: Updating Lua scripts requires new SHA in `scripts.go:53`
- **Connection leaks**: Always close connections with Raw methods
- **Context cancellation**: Handle context timeouts gracefully
- **NewRelic overhead**: Monitor performance impact when enabled

---

*This guide focuses on practical development with go-cache. For complete standards, see [AGENTS.md](AGENTS.md)*
