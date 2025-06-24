# Pool - Fast Worktree Management

Pool provides instant worktree creation by maintaining a pool of pre-created worktrees that can be claimed and renamed on demand.

## Features

- **Instant worktree creation** - Pre-seeded pool means no wait time
- **Automatic pool management** - Pool refills itself in the background
- **Smart branch handling** - Automatically detects remote vs local branches
- **Cleanup utilities** - Remove merged, stale, or orphaned worktrees
- **Cross-platform** - Written in Go, works everywhere

## Installation

```bash
go install github.com/mskelton/pool@latest
```

Or build from source:

```bash
make build
```

## Usage

### Quick Start

```bash
# Initialize pool in current repository
pool init

# Create/switch to a worktree for a branch
pool feature-xyz

# Check pool status
pool status
```

### Commands

#### `pool <branch-name>`
Create or switch to a worktree for the specified branch. If a pool worktree is available, it will be used instantly. Otherwise, a new worktree is created.

#### `pool init`
Initialize a worktree pool in the current repository.

Options:
- `--pool-size N` - Set pool size (default: 5)
- `--convert` - Convert existing repo to bare with worktrees
- `--bare <url>` - Clone repository as bare with pool

#### `pool status`
Show the current pool status and active worktrees.

#### `pool refill`
Manually refill the worktree pool. This happens automatically in the background, but can be triggered manually if needed.

#### `pool clean <type>`
Clean up worktrees based on type:
- `orphaned` - Remove orphaned worktrees
- `stale` - Remove worktrees for deleted branches
- `merged` - Remove worktrees for merged branches
- `pool` - Reset pool worktrees to clean state
- `all` - Run all cleanup tasks

Options:
- `--dry-run` - Preview what would be cleaned

## Configuration

### Environment Variables

- `WORKTREE_POOL_SIZE` - Number of pre-seeded worktrees (default: 5)

## How It Works

1. Pool maintains a set of pre-created worktrees in `.worktree-pool/`
2. When you request a new worktree, pool:
   - Takes an available worktree from the pool
   - Checks out your branch
   - Moves the worktree to the correct location
   - Refills the pool in the background
3. This makes worktree creation nearly instant!

## Development

```bash
# Run tests
make test

# Build binary
make build

# Install locally
make install
```

## License

MIT