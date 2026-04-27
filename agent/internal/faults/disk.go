package faults

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TypeDiskFill is the stable string identifier for the disk fill fault.
const TypeDiskFill = "disk_fill"

// defaultFillDir is where filler files are created when no path is
// passed in parameters.
const defaultFillDir = "/tmp/cloudsentinel-faults"

// fillChunk is the buffer size used when writing the filler file.
// 1 MiB strikes a good balance: small enough not to monopolize the
// disk for too long, big enough to make the syscall overhead
// negligible.
const fillChunk = 1024 * 1024

// DiskFill is a Fault that creates a large file on disk and removes
// it when stopped or on timeout.
//
// The file is written in 1 MiB chunks with a small pause between
// chunks so the fault does not block the disk for too long. Writes
// are synced via fsync to make sure the OS commits the bytes to the
// underlying storage rather than just to the page cache.
type DiskFill struct {
	id        string
	sizeMB    int
	directory string
	path      string // computed at Start
	duration  time.Duration
	startedAt time.Time

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
	stopped bool
}

// NewDiskFill builds a DiskFill fault from typed parameters.
//
// Recognized parameters:
//   - size_mb   (int, default 1024)        — file size in megabytes
//   - directory (string, default /tmp/...)  — where to create the file
func NewDiskFill(duration time.Duration, params Params) (*DiskFill, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("disk_fill: duration must be > 0")
	}

	sizeMB := params.IntOr("size_mb", 1024)
	if sizeMB < 1 {
		return nil, fmt.Errorf("disk_fill: size_mb must be >= 1, got %d", sizeMB)
	}
	if sizeMB > 1048576 {
		return nil, fmt.Errorf("disk_fill: size_mb must be <= 1048576 (1 TiB), got %d", sizeMB)
	}

	directory := params.StringOr("directory", defaultFillDir)

	return &DiskFill{
		id:        uuid.NewString(),
		sizeMB:    sizeMB,
		directory: directory,
		duration:  duration,
		startedAt: time.Now(),
	}, nil
}

// ID returns the unique fault identifier.
func (d *DiskFill) ID() string { return d.id }

// Type returns the stable type string.
func (d *DiskFill) Type() string { return TypeDiskFill }

// StartedAt returns the time the fault was started.
func (d *DiskFill) StartedAt() time.Time { return d.startedAt }

// Duration returns the configured fault duration.
func (d *DiskFill) Duration() time.Duration { return d.duration }

// Path returns the absolute path of the filler file. Empty before
// Start has been called.
func (d *DiskFill) Path() string { return d.path }

// Start creates the directory, opens the filler file, and spawns a
// goroutine that progressively writes the configured size.
func (d *DiskFill) Start(ctx context.Context) error {
	if err := os.MkdirAll(d.directory, 0o755); err != nil {
		return fmt.Errorf("disk_fill: mkdir %q: %w", d.directory, err)
	}

	d.path = filepath.Join(d.directory, "fill-"+d.id+".bin")

	// O_CREATE|O_RDWR|O_TRUNC: make a fresh file every time.
	f, err := os.OpenFile(d.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("disk_fill: open %q: %w", d.path, err)
	}

	workerCtx, cancel := context.WithCancel(ctx)
	d.mu.Lock()
	d.cancel = cancel
	d.mu.Unlock()

	d.wg.Add(1)
	go d.fillLoop(workerCtx, f)
	return nil
}

// Stop cancels the fill loop, waits for it to exit, and removes the
// filler file. Safe to call multiple times.
func (d *DiskFill) Stop(_ context.Context) error {
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		return nil
	}
	d.stopped = true
	cancel := d.cancel
	d.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	d.wg.Wait()

	if d.path != "" {
		// Best-effort removal; if the file is already gone, that's fine.
		_ = os.Remove(d.path)
	}
	return nil
}

// fillLoop writes the requested number of MiB to f, in 1 MiB chunks,
// with a small pause between each chunk. Stops early if ctx is done.
// Closes f before returning.
func (d *DiskFill) fillLoop(ctx context.Context, f *os.File) {
	defer d.wg.Done()
	defer f.Close()

	chunk := make([]byte, fillChunk)
	// Fill the chunk with non-zero bytes so a smart filesystem cannot
	// silently turn the file into a sparse file and "save" the writes.
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	written := 0
	for written < d.sizeMB {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if _, err := f.Write(chunk); err != nil {
			return
		}
		written++

		// Pause briefly so we do not monopolize the disk on a slow
		// medium, and so cancellation stays reactive.
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Millisecond):
		}
	}

	// Sync once at the end so the OS commits the bytes to durable
	// storage rather than just to the page cache.
	_ = f.Sync()
}
