package mainserver

import (
	"DistributedFileSystem/protocol"
	"encoding/json"
	"fmt"
	"net"
)

/*
Server Definition
A server has a listener for requests
A file table that maps fileNames to fileinfo struct which contains storage address
A map of storages along with their available memory.
*/

type MainServer struct {
	listener  net.Listener
	FileTable *FileTable
	Storage   *StorageList
}

func NewMainServer(addr string, storagelist []string) (*MainServer, error) {
	listener, err := net.Listen("tcp", addr)
	fmt.Println("Established Listener at address: ", addr)
	if err != nil {
		fmt.Println("Main Server Create Failed", err)
		return nil, err
	}
	return &MainServer{listener, NewFileTable(), NewStorage(storagelist)}, nil
}

func (ms *MainServer) Start() {
	for {
		conn, err := ms.listener.Accept()
		if err != nil {
			fmt.Println("Main Server Accept Error:", err)
			continue
		}
		ms.handleConnection(conn)
	}
}

func (ms *MainServer) FindStorage(reqMem int64) (StorageAddr string) {
	ms.Storage.lock.RLock()
	defer ms.Storage.lock.RUnlock()

	// Find the largest storage
	maxMem := int64(0)
	var maxAddr string
	for addr, mem := range ms.Storage.nodes {
		if maxMem < mem {
			maxMem = mem
			maxAddr = addr
		}
	}
	// Check if it can store the file
	if reqMem > maxMem {
		return ""
	} else {
		return maxAddr
	}
}

func (ms *MainServer) DeleteRequest(address string, filename string) (success bool) {
	// Establish Connection with storage server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Setup Encoder Decoder
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Build Request
	req := protocol.Delete_Request{Filename: filename}
	payload, err := json.Marshal(req)
	if err != nil {
		return false
	}

	// Send Request
	if err := encoder.Encode(protocol.Message{Type: protocol.DeleteReqM, Payload: payload}); err != nil {
		return false
	}

	// Wait For Response
	var resp protocol.Message
	if err := decoder.Decode(&resp); err != nil {
		return false
	}

	// Unmarshal Response
	var deleteResp protocol.Delete_Response
	if err := json.Unmarshal(resp.Payload, &deleteResp); err != nil {
		return false
	}

	return deleteResp.Success
}

func (ms *MainServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	var msg protocol.Message
	err := decoder.Decode(&msg)
	if err != nil {
		fmt.Println("Main Server Decode Error:", err)
		return
	}

	switch msg.Type {
	case protocol.UploadReq:
		var request protocol.Upload_Request
		if err := json.Unmarshal(msg.Payload, &request); err != nil {
			fmt.Println("Main Server Decode Error:", err)
			return
		}
		fmt.Println("Received Upload Request of file", request.Filename, "with size", request.Size)

		// Check Duplication
		file, exists := ms.Storage.nodes[request.Filename]

		// Allocate StorageList
		storageaddr := ms.FindStorage(request.Size)

		if exists {
			fmt.Println("Main Server File Exists: ", file)
			storageaddr = ""
		} else {
			fmt.Println("Storage Address:", storageaddr)
			ms.FileTable.AddFile(request.Filename, protocol.Fileinfo{
				Filename: request.Filename,
				Size:     request.Size,
				Location: storageaddr,
			})
			ms.Storage.ChangeMem(storageaddr, -request.Size)
		}

		// Build Response
		resp := protocol.Upload_Response{StorageAddr: storageaddr}
		payload, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Main Server Marshal Error:", err)
			return
		}

		// Send Response to Client
		err = encoder.Encode(protocol.Message{Type: protocol.UploadResp, Payload: payload})
		if err != nil {
			fmt.Println("Main Server Encode Error:", err)
			return
		}

	case protocol.DownloadReq:
		var request protocol.Download_Request
		if err := json.Unmarshal(msg.Payload, &request); err != nil {
			fmt.Println("Main Server Decode Error:", err)
			return
		}

		// Look for file
		file, exists := ms.FileTable.GetFile(request.Filename)
		fmt.Println("Received Download Request of file", request.Filename, "Found?", exists)
		var addr string
		if !exists {
			addr = ""
		} else {
			addr = file.Location
			fmt.Println("Storage Address:", addr)
		}

		// Build Response
		resp := protocol.Download_Response{StorageAddr: addr}
		payload, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Main Server Marshal Error:", err)
			return
		}

		// Encode Response
		err = encoder.Encode(protocol.Message{Type: protocol.DownloadResp, Payload: payload})
		if err != nil {
			fmt.Println("Main Server Encode Error:", err)
			return
		}

	case protocol.DeleteReqC:
		var request protocol.Delete_Request
		if err := json.Unmarshal(msg.Payload, &request); err != nil {
			fmt.Println("Main Server Decode Error:", err)
			return
		}

		// Look for file
		file, exists := ms.FileTable.GetFile(request.Filename)
		mem := file.Size
		fmt.Println("Received Delete Request of file", request.Filename, "Found?", exists)
		var addr string
		if !exists {
			// Build Response
			resp := protocol.Delete_Response{Success: false}
			payload, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("Main Server Marshal Error:", err)
				return
			}

			// Encode Response
			err = encoder.Encode(protocol.Message{Type: protocol.DeleteAckM, Payload: payload})
			if err != nil {
				fmt.Println("Main Server Encode Error:", err)
				return
			}
		} else {
			addr = file.Location
			fmt.Println("Location:", addr)

			// Remove File From Table
			ms.FileTable.RemoveFile(request.Filename)

			ms.Storage.ChangeMem(addr, +mem)

			// Notify StorageList Server
			success := ms.DeleteRequest(addr, request.Filename)

			// Build Response
			resp := protocol.Delete_Response{Success: success}
			payload, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("Main Server Marshal Error:", err)
				return
			}

			// Encode Response
			err = encoder.Encode(protocol.Message{Type: protocol.DeleteAckM, Payload: payload})
			if err != nil {
				fmt.Println("Main Server Encode Error:", err)
				return
			}

			ms.FileTable.RemoveFile(request.Filename)

			fmt.Println("Deletion Successful")
		}

	case protocol.LookupReq:
		resp := protocol.Lookup_Response{Files: ms.FileTable.files}
		fmt.Println("Lookup Request Received")
		payload, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Main Server Marshal Error:", err)
			return
		}
		err = encoder.Encode(protocol.Message{Type: protocol.LookupResp, Payload: payload})
		if err != nil {
			fmt.Println("Main Server Encode Error:", err)
			return
		}
	}
}
