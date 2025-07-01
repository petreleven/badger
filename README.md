Here’s a simplified and more direct version of your README. It's structured to be easier to scan and follow, especially for new users who want to get up and running quickly.

---

# 🦡 BadgerWorker

A distributed cron job runner in Go using Redis. Ideal for running scheduled tasks across multiple machines or processes without conflicts.

Inspired by [brooce](https://github.com/SergeyTsalkov/brooce).

---

## ⚡ Quick Start

```bash
git clone https://github.com/petreleven/badger.git
cd badger
go build -o badger badger.go
./badger   # generates default config
```

### Edit Config

Edit `/home/.badger/config.json` to:

* Add your own queues
* Set Redis connection info (default: `redis://localhost:6379`)

### Run the Worker

```bash
./badger       # Run in foreground
./badger -d    # Run as background daemon
```

---

## 🔧 Config Example

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

Each queue must define:

* `Concurrency`: Max jobs in parallel
* `Timeout`: Seconds before a job is considered failed
* `DoneLog`: (optional) Log completed jobs

---

## 🧠 How It Works

* Jobs are defined as shell commands.
* Worker pulls tasks from Redis and runs them.
* Jobs are distributed across all active workers.

To add jobs manually via Redis:

```bash
LPUSH badger:pending:mail-queue task1000
HSET mail-queue task1000 '{"Name":"task1000", "Minute":"*", "Hour":"*", "Day":"01", "Month":"07", "DayWeek":"2", "Job":"uname"}'
```

---

## 🤹 Multiple Workers

**Same Machine**

```bash
./badger
./badger
./badger -d
```

**Across Multiple Machines**

* Use the same `ClusterName` and Redis URL in config
* Workers will auto-distribute load

---

## 💻 Web UI

Starts automatically at:
**[http://localhost:5000](http://localhost:5000)**

### Features:

* View running/completed/failed jobs
* Monitor queue health
* Add/schedule new jobs
* Retry or cancel jobs
* See logs for finished jobs

📸 UI Screenshots
*(Place your images here)*
`./ui_images/ui_home.png`
`./ui_images/queue.png`
`./ui_images/job_logs.png`

---

## 🧾 CLI Options

```bash
./badger -d    # Run as daemon
```

---

## 📜 License

[MIT License](LICENSE)

---

Let me know if you want a separate `USAGE.md` or Docker instructions added.
