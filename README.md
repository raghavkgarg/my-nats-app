# My NATS and MongoDB App

This project is a demonstration of a simple message processing pipeline using NATS for real-time messaging and MongoDB for data persistence. It follows a publish-subscribe (pub-sub) pattern.

## Core Components

*   `publisher.go`: Reads messages line-by-line from `input.txt` and publishes them to a NATS subject.
*   `subscriber.go`: Subscribes to the NATS subject, processes messages, and stores them in a MongoDB collection.
*   `viewer.go`: A utility to view the records stored in MongoDB.
*   `delete.go`: A command-line utility to delete records from MongoDB by `ledger_code`.

## Prerequisites

*   Go (version 1.21 or newer)
*   A running NATS server
*   A running MongoDB instance

## Configuration

The application is configured using environment variables. The following variables are supported, with their default values shown below:

| Environment Variable | Description                       | Default Value                     |
| -------------------- | --------------------------------- | --------------------------------- |
| `NATS_URL`           | The URL of the NATS server.       | `nats://127.0.0.1:4222`           |
| `NATS_SUBJECT`       | The NATS subject for messages.    | `updates`                         |
| `MONGO_URI`          | The connection URI for MongoDB.   | `mongodb://localhost:27017`       |
| `MONGO_DATABASE`     | The database name to use.         | `nats_data`                       |
| `MONGO_COLLECTION`   | The collection name to use.       | `messages`                        |

## How to Run

First, ensure all Go dependencies are downloaded:
```sh
go mod tidy
```

Then, run each component in a separate terminal from the project root directory.

### 1. Run the Subscriber

The subscriber must be running first to receive messages. It will wait for messages or shut down after a timeout or a `Ctrl+C` signal.
```sh
go run subscriber.go models.go config.go
```

### 2. Run the Publisher

This will read `input.txt` and send its contents to the subscriber via NATS.
```sh
go run publisher.go config.go
```

### 3. View Stored Data

This utility queries MongoDB and prints the documents it finds.
```sh
go run viewer.go models.go config.go
```

### 4. Delete Data

This utility deletes documents from MongoDB based on a `ledger_code` provided as a command-line argument. The example below deletes all documents where `ledger_code` is 234.
```sh
go run delete.go config.go 234
```