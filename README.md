# badgetWorker

A cron worker implementation in Go that uses Redis for distributed job queuing and locking. This worker is ideal for environments where multiple instances need to coordinate task execution without stepping on each other's toes.

Inspired by [brooce](https://github.com/SergeyTsalkov/brooce).

## 🚀 Features

- **Simple config-driven setup** – Minimal CLI flags, just a config file and optional daemon mode.
- **Distributed concurrency** – Multiple workers handle jobs cooperatively.
- **Redis-powered** – Uses Redis for queueing and distributed locking.
- **Daemon mode** – Run the worker in the background with the `-d` flag.

## 📦 Installation

Clone the repository and build the binary:

```sh
git clone https://github.com/your-repo/badgetWorker.git
cd badgetWorker
go build -o worker
```

## ⚙️ Configuration

The worker is configured via a JSON file. Here’s an example:

```json
{
  "LogFileName": "badger-33920:3f93b037-924b-451c-abc1-9766e1d2b923.log",
  "PidFileName": "badger-33920:3f93b037-924b-451c-abc1-9766e1d2b923.pid",
  "ClusterName": "badger:allworkers",
  "CustomQueues": {
    "Queues": {
      "mycustomqueue": {
        "Concurrency": 5,
        "Timeout": 60
      }
    }
  },
  "RedisURL": "redis://default:RZiHGPQANRuUdhHoGR@glocalhost:6739"
}
```

### ✏️ Customize Your Queues

Update the `CustomQueues` section to define your own queues. Each queue must have:
- A unique name
- `Concurrency`: number of parallel workers for that queue
- `Timeout`: how long to wait before considering the job failed

## 🧠 Usage

Run the worker normally:

```sh
./worker
```

Or run it as a background daemon:

```sh
./worker -d
```

### 🛠️ CLI Options

```sh
Usage of ./worker:
  -d    Run as a daemon in the background
```

## ✅ Example Queues

```json
"CustomQueues": {
  "Queues": {
    "emailQueue": {
      "Concurrency": 10,
      "Timeout": 30
    },
    "pdfGeneration": {
      "Concurrency": 3,
      "Timeout": 120
    }
  }
}
```

## 📝 License

[MIT License](LICENSE)
```
