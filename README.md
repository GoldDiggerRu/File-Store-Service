# File Store Service

A simple HTTP server written in Go for storing and managing files.

## Features

- **Health Check** - `GET /health` - Check if the service is running
- **Upload Files** - `POST /upload` - Upload files to the server
- **List Files** - `GET /list` - List all stored files

## Installation

### Prerequisites

You need to have Go installed on your system.

#### On Windows:

1. Download Go from https://golang.org/dl/
2. Select the Windows installer (MSI) appropriate for your system (amd64 or arm64)
3. Run the installer and follow the installation wizard
4. Verify installation by opening PowerShell and running:
   ```powershell
   go version
   ```

#### On macOS:

1. Download the macOS installer from https://golang.org/dl/
2. Run the installer and follow the prompts
3. Verify installation:
   ```bash
   go version
   ```

#### On Linux:

```bash
# Download and extract (replace VERSION with current Go version)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

## Running the Server

### Option 1: Direct Execution

```powershell
# Navigate to the project directory
cd .\file-store-service

# Run the server
go run main.go
```

The server will start on `http://localhost:8080`

### Option 2: Build and Run

```powershell
# Build the executable
go build -o file-store-service.exe

# Run the executable
.\file-store-service.exe
```

## Testing the Service

Once the server is running, you can test the endpoints:

### Health Check
```powershell
curl http://localhost:8080/health
```

### Upload a File
```powershell
# Create a test file
echo "Hello, File Store Service!" > test.txt

# Upload it
curl -X POST -F "file=@test.txt" http://localhost:8080/upload
```

### List Files
```powershell
curl http://localhost:8080/list
```

## Project Structure

```
file-store-service/
├── main.go          # Main application file
├── go.mod           # Go module definition
└── README.md        # This file
```

## Storage

Uploaded files are stored in the `file_storage/` directory, which is created automatically when the server starts.

## Building for Different Platforms

```powershell
# For Windows
go build -o file-store-service.exe

# For Linux
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o file-store-service

# For macOS
$env:GOOS = "darwin"; $env:GOARCH = "amd64"; go build -o file-store-service
```
