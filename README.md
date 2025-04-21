# Distributed-File-System

An implementation of a scalable and modular distributed file system in Go.

This document explains how to use the command-line interface for a distributed file system with three roles: main server, storage server, and client.

## Overview of Roles

- **Main Server**: Coordinates between clients and storage servers, managing file metadata.  
- **Storage Server**: Stores and retrieves files.  
- **Client**: Allows users to upload, download, delete, or lookup files.

## Shared Flags

- `-role <role>`: Specifies the role (`"main"`, `"storage"`, or `"client"`).  
- `-listen_addr <address>`: Address to listen on (required for main and storage servers).

---

## Running the System

To operate the system:

### Start Storage Servers

```bash
go run main.go -role storage -listen_addr <storage_address> -storage_dir <directory> [-available_mem <memory_in_bytes>]
```

**Example:**

```bash
go run main.go -role storage -listen_addr localhost:8081 -storage_dir ./StorageNode1 -available_mem 1000000000
```

### Start the Main Server

```bash
go run main.go -role main -listen_addr <main_address> -storage_addrs <storage_addresses>
```

**Example:**

```bash
go run main.go -role main -listen_addr localhost:8080 -storage_addrs localhost:8081,localhost:8082
```

### Use the Client

```bash
go run main.go -role client -main_addr <main_address> -cmd <command> [additional flags]
```

---

## Role-Specific Usage

### 1. Main Server

**Description**: Manages client requests and file metadata.

**Required Flags:**

- `-role main`  
- `-listen_addr <address>`: Listening address (e.g., `"localhost:8080"`)  
- `-storage_addrs <addresses>`: Comma-separated storage server addresses (e.g., `"localhost:8081,localhost:8082"`)

**Example:**

```bash
go run main.go -role main -listen_addr localhost:8080 -storage_addrs localhost:8081,localhost:8082
```

---

### 2. Storage Server

**Description**: Handles file storage and retrieval.

**Required Flags:**

- `-role storage`  
- `-listen_addr <address>`: Listening address (e.g., `"localhost:8081"`)  
- `-storage_dir <directory>`: Storage directory (e.g., `"./StorageNode1"`)

**Optional Flags:**

- `-available_mem <memory_in_bytes>`: Memory limit in bytes (default: `-1`, unlimited)

**Example:**

```bash
go run main.go -role storage -listen_addr localhost:8081 -storage_dir ./StorageNode1 -available_mem 1000000000
```

---

### 3. Client

**Description**: Interacts with the system for file operations.

**Required Flags:**

- `-role client`  
- `-main_addr <address>`: Main server address (e.g., `"localhost:8080"`)  
- `-cmd <command>`: Command (`"upload"`, `"download"`, `"delete"`, `"lookup"`)

**Additional Flags:**

- `-filename <filename>`: File for upload, download, or delete  
- `-output <output_filename>`: Local file for download

**Examples:**

#### Upload

```bash
go run main.go -role client -main_addr localhost:8080 -cmd upload -filename test.txt
```

#### Download

```bash
go run main.go -role client -main_addr localhost:8080 -cmd download -filename test.txt -output downloaded.txt
```

#### Delete

```bash
go run main.go -role client -main_addr localhost:8080 -cmd delete -filename test.txt
```

#### Lookup

```bash
go run main.go -role client -main_addr localhost:8080 -cmd lookup
```

---

## Notes

1. **Start the servers in the following order**:  
   `Storage` → `Main` → `Client`

2. **Metadata is not persistent**:  
   Shutting down the main server will remove all metadata.
