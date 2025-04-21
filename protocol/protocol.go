package protocol

import (
	"encoding/json"
)

type MessageType string

const (
	UploadReq   MessageType = "CLIENT_UPLOAD_REQ"
	DeleteReqC  MessageType = "CLIENT_DELETE_REQ"
	DownloadReq MessageType = "CLIENT_DOWNLOAD_REQ"
	LookupReq   MessageType = "CLIENT_LOOKUP_REQ"

	UploadResp   MessageType = "MAIN_UPLOAD_RESP"
	DownloadResp MessageType = "MAIN_DOWNLOAD_RESP"
	DeleteReqM   MessageType = "MAIN_DELETE_REQ"
	DeleteAckM   MessageType = "MAIN_DELETE_ACK"
	LookupResp   MessageType = "MAIN_LOOKUP_RESP"
	MemLookupReq MessageType = "MAIN_MEM_LOOKUP_REQ"

	UploadAck     MessageType = "NODE_UPLOAD_ACK"
	DownloadAck   MessageType = "NODE_DOWNLOAD_ACK"
	DeleteAckN    MessageType = "NODE_DELETE_ACK"
	MemLookupResp MessageType = "NODE_MEM_LOOKUP_RESP"

	Error MessageType = "ERROR"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

/*
Upload Process
Client -> Main for allocation
Main -> Client for address
Client -> Node for upload
Node -> Client for confirmation
*/

// Client Upload Request
type Upload_Request struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// Main Server Upload Response
type Upload_Response struct {
	StorageAddr string `json:"storage_addr"`
}

/*
Download Process
Client -> Main for request
Main -> Client for address
Client -> Node for request
Node -> Client for download
*/
type Download_Request struct {
	Filename string `json:"filename"`
}

type Download_Response struct {
	StorageAddr string `json:"storage_addr"`
}

/*
Delete Process
Client -> Main for request
Main -> Node for deletion
Node -> Main for confirmation
Main -> Client for confirmation
*/

// Client/Main Delete Request
type Delete_Request struct {
	Filename string `json:"filename"`
}

// Main/Node Delete Response
type Delete_Response struct {
	Success bool `json:"success"`
}

/*
Lookup Process
Client -> Main for request
Main -> Client for result
*/

type Fileinfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Location string `json:"location"`
}

// Main Lookup Response
type Lookup_Response struct {
	Files map[string]Fileinfo `json:"files"`
}

// Mem Lookup Response
type MemLookup_Response struct {
	Availmem int64 `json:"availmem"`
}
