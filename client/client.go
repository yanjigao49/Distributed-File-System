package client

import (
	"DistributedFileSystem/protocol"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	mainAddress string
}

func NewClient(mainAddress string) *Client {
	return &Client{mainAddress: mainAddress}
}

func (c *Client) Upload(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file status
	fileinfo, err := file.Stat()
	if err != nil {
		return err
	}
	fmt.Println("Uploading", fileinfo.Name(), ", Size:", fileinfo.Size())

	// Connect to Main Server
	conn, err := net.Dial("tcp", c.mainAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Construct Request
	req := protocol.Upload_Request{
		Filename: filename,
		Size:     fileinfo.Size(),
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Create encoder and decoder
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Send message
	err = encoder.Encode(protocol.Message{
		Type:    protocol.UploadReq,
		Payload: payload,
	})
	if err != nil {
		return err
	}

	// Receive message
	var msg protocol.Message
	err = decoder.Decode(&msg)
	if err != nil {
		return err
	}
	if msg.Type != protocol.UploadResp {
		return fmt.Errorf("UploadResp expected")
	}
	var resp protocol.Upload_Response

	err = json.Unmarshal(msg.Payload, &resp)
	if err != nil {
		return err
	}

	if resp.StorageAddr == "" {
		return fmt.Errorf("No Storage Available or File Already Exists")
	}

	// Connect to storage server
	storageConn, err := net.Dial("tcp", resp.StorageAddr)
	if err != nil {
		return err
	}
	defer storageConn.Close()

	// Send file data
	encoder = json.NewEncoder(storageConn)
	decoder = json.NewDecoder(storageConn)
	err = encoder.Encode(protocol.Message{
		Type:    protocol.UploadReq,
		Payload: payload,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(&msg)
	if err != nil {
		return err
	}

	if msg.Type != protocol.UploadAck {
		return fmt.Errorf("UploadAck expected")
	} else {
		// Send file data
		_, err = io.Copy(storageConn, file)
		return err
	}
}

func (c *Client) Download(filename string, outputpath string) error {
	// Request storage address from main server
	conn, err := net.Dial("tcp", c.mainAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Build Request
	req := protocol.Download_Request{
		Filename: filename,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Send Request
	err = encoder.Encode(protocol.Message{
		Type:    protocol.DownloadReq,
		Payload: payload,
	})

	if err != nil {
		return err
	}

	var msg protocol.Message
	err = decoder.Decode(&msg)
	if err != nil {
		return err
	}
	if msg.Type != protocol.DownloadResp {
		return fmt.Errorf("DownloadResp expected")
	}
	var resp protocol.Download_Response
	err = json.Unmarshal(msg.Payload, &resp)
	if err != nil {
		return err
	}
	if resp.StorageAddr == "" {
		return fmt.Errorf("File Not Found")
	}

	// Connect to storage server
	storageConn, err := net.Dial("tcp", resp.StorageAddr)
	if err != nil {
		return err
	}

	// Send download request
	defer storageConn.Close()
	encoder = json.NewEncoder(storageConn)
	decoder = json.NewDecoder(storageConn)
	err = encoder.Encode(protocol.Message{
		Type:    protocol.DownloadReq,
		Payload: payload,
	})

	err = decoder.Decode(&msg)
	if err != nil {
		return err
	}

	if msg.Type != protocol.DownloadAck {
		return fmt.Errorf("DownloadAck expected")
	} else {
		// Save file
		f, err := os.Create(outputpath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, storageConn)
		return err
	}

}

func (c *Client) Delete(filename string) (bool, error) {
	conn, err := net.Dial("tcp", c.mainAddress)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	req := protocol.Delete_Request{
		Filename: filename,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return false, err
	}
	err = encoder.Encode(protocol.Message{
		Type:    protocol.DeleteReqC,
		Payload: payload,
	})
	if err != nil {
		return false, err
	}

	// Receive response
	var msg protocol.Message
	err = decoder.Decode(&msg)
	if err != nil {
		return false, err
	}
	if msg.Type != protocol.DeleteAckM {
		return false, fmt.Errorf("DeleteAckM expected")
	}
	var resp protocol.Delete_Response
	err = json.Unmarshal(msg.Payload, &resp)
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func (c *Client) Lookup() (map[string]protocol.Fileinfo, error) {
	conn, err := net.Dial("tcp", c.mainAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	err = encoder.Encode(protocol.Message{
		Type: protocol.LookupReq,
	})
	if err != nil {
		return nil, err
	}

	var msg protocol.Message
	err = decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}
	if msg.Type != protocol.LookupResp {
		return nil, fmt.Errorf("LookupResp expected")
	}
	var resp protocol.Lookup_Response
	err = json.Unmarshal(msg.Payload, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Files, nil
}
