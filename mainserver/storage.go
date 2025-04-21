package mainserver

import (
	"DistributedFileSystem/protocol"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type StorageList struct {
	lock  sync.RWMutex
	nodes map[string]int64 // Address -> Available Mem
}

func getAvailableMemory(address string) int64 {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return -1
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Ask for memory on server
	err = encoder.Encode(protocol.Message{
		Type: protocol.MemLookupReq,
	})
	if err != nil {
		return -1
	}

	// Get Response
	var msg protocol.Message
	err = decoder.Decode(&msg)
	if err != nil {
		return -1
	}
	if msg.Type != protocol.MemLookupResp {
		return -1
	}

	var resp protocol.MemLookup_Response
	err = json.Unmarshal(msg.Payload, &resp)
	if err != nil {
		return -1
	}

	return resp.Availmem
}

func NewStorage(storagelist []string) *StorageList {
	nodes := make(map[string]int64)
	for _, addr := range storagelist {
		fmt.Println("Establishing connection with storage server at address:", addr)
		nodes[addr] = getAvailableMemory(addr)
		fmt.Println("Server", addr, "is available with memory:", nodes[addr])
	}

	return &StorageList{
		nodes: nodes,
	}
}

func (ft *StorageList) Add(address string, mem int64) {
	ft.lock.Lock()
	ft.nodes[address] = mem
	ft.lock.Unlock()
}

func (ft *StorageList) Remove(address string) {
	ft.lock.Lock()
	delete(ft.nodes, address)
	ft.lock.Unlock()
}

func (ft *StorageList) ChangeMem(address string, amt int64) {
	ft.lock.Lock()
	ft.nodes[address] += amt
	ft.lock.Unlock()
}
