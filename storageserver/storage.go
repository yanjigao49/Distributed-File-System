package storageserver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type Storage struct {
	lock      sync.RWMutex
	path      string
	capacity  int64
	available int64
	fileLocks map[string]*sync.RWMutex
}

func NewStorage(path string, mem int64) *Storage {
	return &Storage{
		path:      path,
		fileLocks: make(map[string]*sync.RWMutex),
		capacity:  mem,
		available: mem,
	}
}

func (storage *Storage) getLock(key string) *sync.RWMutex {
	storage.lock.RLock()
	defer storage.lock.RUnlock()

	// Ensure Non Empty lock is returned
	if _, ok := storage.fileLocks[key]; !ok {
		storage.fileLocks[key] = &sync.RWMutex{}
	}
	return storage.fileLocks[key]
}

func (storage *Storage) getCapacity() int64 {
	return storage.capacity
}

func (storage *Storage) getAvailableMemory() int64 {
	return storage.available
}

func (storage *Storage) Upload(filename string, size int64, reader io.Reader) error {
	fileLock := storage.getLock(filename)
	fileLock.Lock()
	defer fileLock.Unlock()

	path := filepath.Join(storage.path, filename)

	var prevSize int64 = 0
	if fileinfo, err := os.Stat(path); err == nil {
		prevSize = fileinfo.Size()
	}

	storage.lock.RLock()
	if size-prevSize > storage.available {
		fmt.Errorf("Not enough space to upload to %s", filename)
	}
	storage.lock.RUnlock()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	written, err := io.CopyN(file, reader, size)
	if err != nil {
		return err
	}

	storage.lock.Lock()
	storage.available -= (written - prevSize)
	fmt.Println("Upload Successful, Available Memory:", storage.available)
	storage.lock.Unlock()

	return err
}

func (storage *Storage) Download(filename string, writer io.Writer) error {
	fileLock := storage.getLock(filename)
	fileLock.RLock()
	defer fileLock.RUnlock()
	path := filepath.Join(storage.path, filename)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}

func (storage *Storage) Delete(filename string) error {
	fileLock := storage.getLock(filename)
	fileLock.Lock()
	defer fileLock.Unlock()
	path := filepath.Join(storage.path, filename)

	// Get Size
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	size := info.Size()

	// Perform Deletion
	err = os.Remove(path)

	if err != nil {
		return err
	}

	storage.lock.Lock()
	storage.available += size
	storage.lock.Unlock()

	return err
}
