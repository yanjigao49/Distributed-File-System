package main

import (
	"DistributedFileSystem/client"
	"DistributedFileSystem/mainserver"
	"DistributedFileSystem/storageserver"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	role := flag.String("role", "", "Role to use: main, storage, client")

	// Shared Args
	listenaddr := flag.String("listen_addr", "", "Address to listen on")

	// Main Server Args
	storageaddrs := flag.String("storage_addrs", "", "Storage addresses, comma separated") //localhost:8081,localhost:8082 ...

	// Storage Server Args
	storagedir := flag.String("storage_dir", "", "Directory to store file") // ./StorageNode1 ...
	availablemem := flag.Int64("available_mem", -1, "Available memory")

	// Client Args
	mainaddr := flag.String("main_addr", "", "Main server address")
	command := flag.String("cmd", "", "Command to execute: upload, download, delete, lookup")
	filename := flag.String("filename", "", "Filename to upload/download/delete")
	output := flag.String("output", "", "Output filename for download")

	flag.Parse()

	switch *role {
	case "main":
		storageList := strings.Split(*storageaddrs, ",")
		server, err := mainserver.NewMainServer(*listenaddr, storageList)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Main server listening on", *listenaddr)
		server.Start()

	case "storage":
		if *availablemem < 0 {
			fmt.Println("Available memory is required")
		}
		if err := os.MkdirAll(*storagedir, 0755); err != nil {
			fmt.Println("Stroage dir error", err)
			os.Exit(1)
		}

		server, err := storageserver.NewStorageServer(*listenaddr, *storagedir, *availablemem)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Storage server listening on", *listenaddr)
		server.Start()

	case "client":
		client := client.NewClient(*mainaddr)
		switch *command {
		case "upload":
			if *filename == "" {
				fmt.Println("Filename is required")
				os.Exit(1)
			}
			err := client.Upload(*filename)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("Upload successful")
		case "download":
			if *filename == "" || *output == "" {
				fmt.Println("Filename and Output is required")
				os.Exit(1)
			}
			if err := client.Download(*filename, *output); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("Download successful")
		case "delete":
			if *filename == "" {
				fmt.Println("Filename is required")
				os.Exit(1)
			}
			success, err := client.Delete(*filename)
			if err != nil || !success {
				fmt.Println("Deletion Failed")
				os.Exit(1)
			}
			fmt.Println("Deletion successful")
		case "lookup":
			files, err := client.Lookup()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, file := range files {
				fmt.Println("Filename:", file.Filename, "Size:", file.Size)
			}
		default:
			fmt.Println("Invalid command")
			os.Exit(1)
		}

	default:
		fmt.Println("Invalid role")
		os.Exit(1)
	}
}

func splitByComma(input string) []string {
	if input == "" {
		return nil
	}
	return strings.Split(input, ",")
}
