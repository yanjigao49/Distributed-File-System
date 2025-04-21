package mainserver

import (
	"DistributedFileSystem/protocol"
	"sync"
)

type FileTable struct {
	lock  sync.RWMutex
	files map[string]protocol.Fileinfo
}

func NewFileTable() *FileTable {
	return &FileTable{
		files: make(map[string]protocol.Fileinfo),
	}
}

func (ft *FileTable) AddFile(filename string, file protocol.Fileinfo) {
	ft.lock.Lock()
	ft.files[filename] = file
	ft.lock.Unlock()
}

func (ft *FileTable) RemoveFile(filename string) bool {
	ft.lock.Lock()
	defer ft.lock.Unlock()
	_, exists := ft.files[filename]
	if exists {
		delete(ft.files, filename)
		return true
	}
	return false
}

func (ft *FileTable) GetFile(filename string) (protocol.Fileinfo, bool) {
	ft.lock.RLock()
	defer ft.lock.RUnlock()
	file, exists := ft.files[filename]
	return file, exists
}

func (ft *FileTable) ListFiles() []protocol.Fileinfo {
	ft.lock.RLock()
	defer ft.lock.RUnlock()
	files := make([]protocol.Fileinfo, 0, len(ft.files))
	for _, file := range ft.files {
		files = append(files, file)
	}
	return files
}
