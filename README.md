# badgetWorker

A cron worker implementation in Go that utilizes Redis as a data queue and for distributed locking. This project is heavily inspired by [brooce](https://github.com/SergeyTsalkov/brooce).

You can run multiple instances of `badgetWorker`, and they will coordinate job execution using Redis, ensuring efficient distributed scheduling.

## Features

- **Distributed job handling** – Multiple workers can run concurrently, preventing job duplication.
- **Redis-based coordination** – Uses Redis for job queuing and locking.
- **Daemon mode** – Supports running as a background process.

## Installation

Clone the repository and build the binary:

```sh
git clone https://github.com/your-repo/badgetWorker.git
cd badgetWorker
go build -o worker
```

## Usage

Run the worker with the following command:

```sh
./worker -d=0 -u=redis://localhost:6739 -name=cronworker
```

### Command-line options

```sh
Usage of ./worker:
  -d int
        Run as a daemon (0 by default, 1 to run in daemon mode)
  -name string
        Name of the worker instance (default: "worker1")
  -u string
        Redis URL (default: "redis://localhost:6739")
```

## Example

To start a worker instance named `cronworker` and connect to Redis at `localhost:6739`:

```sh
./worker -d=0 -u=redis://localhost:6739 -name=cronworker
```

To run in daemon mode:

```sh
./worker -d=1 -u=redis://localhost:6739 -name=backgroundworker
```

## License

[MIT License](LICENSE)

