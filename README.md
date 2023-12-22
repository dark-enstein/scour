# Scour

Scour is a command-line tool developed in Go, offering functionalities similar to `curl`. It allows users to make HTTP requests and interact with UNIX sockets.

## Features

- Supports HTTP methods: GET, POST, PUT, DELETE.
- Ability to pass custom request headers and data.
- Supports verbose output for debugging.
- Can connect through an abstract Unix domain socket.

## Installation

Clone the repository and build the project:

```bash
    git clone https://github.com/dark-enstein/scour.git
    cd scour
    make build
```

## Usage

### Basic HTTP Request
```bash
    scour [flags] <url>
```

Flags:
- `--verbose` or `-v`: Enable verbose mode.
- `-X`: Specify the request method (GET, POST, etc.).
- `-d`: Pass request data.
- `-H`: Custom request headers.
- `--abstract-unix-socket` or `-aus`: Use an abstract Unix domain socket.

### Example
```bash
    scour -v -X GET https://example.com
```


## Docker Support

Build a Docker image using the provided Dockerfile:

```bash
    make docker-build
```


## Testing

Run tests with:

```bash
    make test
```


## Contributing

Contributions are welcome. Please submit pull requests or open issues for any bugs or feature requests.
