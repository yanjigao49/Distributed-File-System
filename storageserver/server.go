package storageserver

import (
	"DistributedFileSystem/protocol"
	"encoding/json"
	"fmt"
	"net"
)

type StorageServer struct {
	listener *net.TCPListener
	storage  *Storage
}

func (s *StorageServer) GetPath() string {
	return s.storage.path
}

func (s *StorageServer) GetAvailableMemory() int64 {
	return s.storage.available
}

func (s *StorageServer) GetCapacityMemory() int64 {
	return s.storage.capacity
}

func NewStorageServer(addr, dir string, mem int64) (*StorageServer, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return &StorageServer{
		listener: listener,
		storage:  NewStorage(dir, mem),
	}, nil
}

func (s *StorageServer) Start() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Acceptance Error", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *StorageServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var msg protocol.Message
	err := decoder.Decode(&msg)
	if err != nil {
		fmt.Println("Decode Error", err)
		return
	}

	switch msg.Type {
	case protocol.UploadReq:
		var req protocol.Upload_Request
		// Unmarshal payload
		err := json.Unmarshal(msg.Payload, &req)
		if err != nil {
			fmt.Println("Decode Error", err)
			return
		}
		fmt.Println("Received Upload Request with File", req.Filename, "with size", req.Size)

		if s.GetAvailableMemory() < req.Size {
			fmt.Println("Not enough available memory")
			return
		}

		// Send Response
		err = encoder.Encode(protocol.Message{
			Type: protocol.UploadAck,
		})

		if err != nil {
			fmt.Println("Encode Error", err)
			return
		}

		// Receive File Data
		if err := s.storage.Upload(req.Filename, req.Size, conn); err != nil {
			fmt.Println("Upload Error", err)
			return
		}
		fmt.Println("Upload Successful")

	case protocol.DownloadReq:
		var req protocol.Download_Request

		// Unmarshal Request
		err := json.Unmarshal(msg.Payload, &req)
		if err != nil {
			fmt.Println("Decode Error", err)
			return
		}
		fmt.Println("Received Download Request with File", req.Filename)

		// Send Response
		err = encoder.Encode(protocol.Message{
			Type: protocol.DownloadAck,
		})

		if err != nil {
			fmt.Println("Encode Error", err)
			return
		}

		// Perform Download
		if err := s.storage.Download(req.Filename, conn); err != nil {
			fmt.Println("Download Error", err)
			return
		}

	case protocol.DeleteReqM:
		var req protocol.Delete_Request
		fmt.Println("Received Delete Request with File", req.Filename)
		// Unmarshal Request
		err := json.Unmarshal(msg.Payload, &req)
		if err != nil {
			fmt.Println("Unmarshal Error", err)
			return
		}

		payload, err := json.Marshal(protocol.Delete_Response{
			Success: true,
		})
		if err != nil {
			fmt.Println("Marshal Error", err)
			return
		}

		// Send Response
		err = encoder.Encode(protocol.Message{
			Type:    protocol.DeleteAckN,
			Payload: payload,
		})

		if err != nil {
			fmt.Println("Encode Error", err)
			return
		}

		// Perform Deletion
		if err := s.storage.Delete(req.Filename); err != nil {
			fmt.Println("Delete Error", err)
			return
		}
		fmt.Println("Deletion Successful, Available Memory", s.GetAvailableMemory())
	case protocol.MemLookupReq:
		fmt.Println("Received MemLookup request, Current memory:", s.GetAvailableMemory(), ", Capacity:", s.GetCapacityMemory())
		payload, err := json.Marshal(protocol.MemLookup_Response{
			Availmem: s.GetAvailableMemory(),
		})
		if err != nil {
			fmt.Println("Decode Error", err)
			return
		}

		// Send Response
		err = encoder.Encode(protocol.Message{
			Type:    protocol.MemLookupResp,
			Payload: payload,
		})

		if err != nil {
			fmt.Println("Encode Error", err)
			return
		}
	}
}
