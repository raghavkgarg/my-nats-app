# My NATS and MongoDB App

This project is a demonstration of a simple message processing pipeline using NATS for real-time messaging and MongoDB for data persistence. It follows a publish-subscribe (pub-sub) pattern with a modular Go project structure.

## Project Structure

```
my-nats-app/
├── cmd/                    # Application entry points
│   ├── delete/            # Delete utility command
│   │   └── main.go
│   ├── publisher/         # Message publisher command
│   │   ├── main.go
│   │   └── input.txt      # Sample input file
│   ├── subscriber/        # Message subscriber command
│   │   └── main.go
│   ├── viewer/           # Database viewer command
│   │   └── main.go
│   └── webserver/        # Web interface command
│       └── main.go
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   │   └── config.go
│   ├── db/              # Database connections
│   │   └── mongo.go
│   ├── errors/          # Custom error types
│   │   └── errors.go
│   ├── handlers/        # HTTP handlers and web interface
│   │   ├── static/      # Static web assets
│   │   │   ├── index.html
│   │   │   ├── inquiry.html
│   │   │   └── style.css
│   │   └── web.go
│   └── models/          # Data models
│       └── message.go
├── go.mod
├── go.sum
└── README.md
```

## Core Components

* **Publisher** (`cmd/publisher/`): Reads messages line-by-line from an input file and publishes them to a NATS subject.
* **Subscriber** (`cmd/subscriber/`): Subscribes to the NATS subject, processes messages, and stores them in a MongoDB collection.
* **Viewer** (`cmd/viewer/`): A utility to view the records stored in MongoDB.
* **Delete** (`cmd/delete/`): A command-line utility to delete records from MongoDB by `ledger_code`.
* **Web Server** (`cmd/webserver/`): A web server providing a user interface to manually submit data, inquire about existing entries, and delete records by `ledger_code`.

## Prerequisites

* Go (version 1.21 or newer)
* A running NATS server
* A running MongoDB instance

## Configuration

The application is configured using environment variables. The following variables are supported, with their default values shown below:

| Environment Variable | Description                       | Default Value                     |
| -------------------- | --------------------------------- | --------------------------------- |
| `NATS_URL`           | The URL of the NATS server.       | `nats://localhost:4222`           |
| `NATS_SUBJECT`       | The NATS subject for messages.    | `messages`                        |
| `MONGO_URI`          | The connection URI for MongoDB.   | `mongodb://localhost:27017`       |
| `MONGO_DB`           | The database name to use.         | `messagedb`                       |
| `MONGO_COLLECTION`   | The collection name to use.       | `messages`                        |
| `WEB_PORT`           | The port for the web server.      | `8080`                            |

## Installation and Setup

First, ensure all Go dependencies are downloaded:
```sh
go mod tidy
```

## How to Run

Run each component in a separate terminal from the project root directory.

### 1. Run the Subscriber

The subscriber must be running first to receive messages. It will wait for messages or shut down after a timeout or a `Ctrl+C` signal.
```sh
go run cmd/subscriber/main.go
```

### 2. Run the Publisher

This will read the input file and send its contents to the subscriber via NATS. By default, it reads from `cmd/publisher/input.txt`, but you can specify a different file:
```sh
# Using default input file
go run cmd/publisher/main.go

# Using custom input file
go run cmd/publisher/main.go -file /path/to/your/input.txt
```

### 3. View Stored Data

This utility queries MongoDB and prints the documents it finds.
```sh
go run cmd/viewer/main.go
```

### 4. Delete Data

This utility deletes documents from MongoDB based on a `ledger_code` provided as a command-line argument. The example below deletes all documents where `ledger_code` is 234.
```sh
go run cmd/delete/main.go 234
```

### 5. Run the Web Server

Start the web interface to manually submit data and query existing entries:
```sh
go run cmd/webserver/main.go
```

Then open your browser and navigate to `http://localhost:8080` (or the port specified in `WEB_PORT`).

The web interface provides three main pages:
- **Data Entry Page** (`/`): Submit new message data
- **Inquiry Page** (`/inquiry-page`): Search for records by ledger code
- **Delete Page** (`/delete-page`): Delete records by ledger code with confirmation

## Message Format

The application expects messages in a specific format where:
- Characters 1-4: Meter identifier (`accMtr`)
- Characters 5-7: Ledger code (`acc`) - must be a numeric value
- Messages with ledger code `123` are filtered out and not stored

Example input format: `xxxx234` where `xxxx` is the meter ID and `234` is the ledger code.

## Architecture Notes

This project follows Go's standard project layout:
- `cmd/` contains the main applications for this project
- `internal/` contains private application and library code
- Each command is self-contained in its own directory under `cmd/`
- Shared functionality is organized in packages under `internal/`

The modular structure makes it easy to:
- Build individual components separately
- Test components in isolation
- Maintain and extend functionality
- Deploy components independently if needed

## Building Binaries

You can build individual binaries for each component:
```sh
# Build all components
go build -o bin/publisher cmd/publisher/main.go
go build -o bin/subscriber cmd/subscriber/main.go
go build -o bin/viewer cmd/viewer/main.go
go build -o bin/delete cmd/delete/main.go
go build -o bin/webserver cmd/webserver/main.go
```

## Development

The internal packages provide:
- `config`: Centralized configuration management
- `db`: Database connection utilities
- `models`: Data structures and models
- `handlers`: HTTP handlers for the web interface
- `errors`: Custom error types and handling