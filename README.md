---

# 🦡 BadgerWorker

A fast, distributed cron job runner written in Go, backed by Redis. Ideal for running and managing scheduled shell jobs across multiple machines or processes.

Inspired by**: [brooce](https://github.com/SergeyTsalkov/brooce)

---

## ⚡ Quick Start

### 📥 Download Prebuilt Binaries

Get the latest binaries from the [Releases page](https://github.com/petreleven/badger/releases):

| Platform        | Download                                                                                     |
| --------------- | -------------------------------------------------------------------------------------------- |
| Linux (x86\_64) | [badger-linux](https://github.com/petreleven/badger/releases/download/v1.0.0/badger-linux)   |
| macOS (x86\_64) | [badger-darwin](https://github.com/petreleven/badger/releases/download/v1.0.0/badger-darwin) |
| Windows         | [badger.exe](https://github.com/petreleven/badger/releases/download/v1.0.0/badger.exe)       |

Make it executable (Linux/macOS):

```bash
chmod +x ./badger-linux
./badger
```

### 🛠️ Build from Source

```bash
git clone https://github.com/petreleven/badger.git
cd badger
go build -o badger badger.go
./badger   # generates config at /home/.badger/config.json
```

---

## ⚙️ Configuration

After first run, edit `/home/.badger/config.json` to:

* Set Redis connection (default: `redis://localhost:6379`)
* Define queues and job parameters

### Example Config

```json
{
  "RedisURL": "redis://localhost:6379",
  "ClusterName": "badger:allworkers",
  "CustomQueues": {
    "Queues": {
      "mail-queue": {
        "Concurrency": 3,
        "Timeout": 120,
        "DoneLog": true
      }
    }
  }
}
```

### Queue Fields

* `Concurrency`: Max parallel jobs
* `Timeout`: Seconds until a job is considered failed
* `DoneLog` (optional): Log completed jobs

---

## 🚀 Running the Worker

```bash
./badger       # Foreground mode
./badger -d    # Background daemon
```

---

## 🔁 Adding Jobs (Manually via Redis)

Badger executes shell jobs defined in Redis. Example:

```bash
LPUSH badger:pending:mail-queue task1000
HSET mail-queue task1000 '{"Name":"task1000", "Minute":"*", "Hour":"*", "Day":"01", "Month":"07", "DayWeek":"2", "Job":"uname"}'
```

You can add more job objects the same way using different task names.

---

## 🤹 Scale with Multiple Workers

### On the Same Machine

```bash
./badger
./badger
./badger -d
```

### On Different Machines

* Ensure all workers share the same `RedisURL` and `ClusterName`.
* Jobs will be automatically distributed across all workers.

---

## 💻 Web Interface

Accessible at:
**[http://localhost:5000](http://localhost:5000)**

### Web UI Features

* View active/completed/failed jobs
* Inspect or manage queues
* Add and schedule new jobs
* Retry or cancel existing jobs
* Search logs (if `DoneLog: true`)

📸 UI Screenshots:

* `./ui_images/ui_home.png`
* `./ui_images/queue.png`
* `./ui_images/job_logs.png`

---

## 🧾 CLI Options

```bash
./badger -d      # Run in background
```

---

## 📄 License

[MIT License](LICENSE)

---

Let me know if you'd like a `Dockerfile`, systemd service example, or separate `USAGE.md` guide included!
