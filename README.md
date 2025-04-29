# badgerWorker

A cron worker implementation in Go that uses Redis for distributed job queuing and locking. This worker is ideal for environments where multiple instances need to coordinate task execution without stepping on each other's toes.

Inspired by [brooce](https://github.com/SergeyTsalkov/brooce).

## 🚀 Features

- **Simple config-driven setup** – Minimal CLI flags, just a config file and optional daemon mode.
- **Distributed concurrency** – Multiple workers handle jobs cooperatively.
- **Redis-powered** – Uses Redis for queueing and distributed locking.
- **Daemon mode** – Run the worker in the background with the `-d` flag.
- **Web UI** – Monitor and manage jobs through a browser-based interface.

## 📦 Installation

Clone the repository and build the binary:

```sh
git clone https://github.com/your-repo/badger.git
cd badger
go build -o badger  badger.go
```

## ⚙️ Configuration

The worker is configured via a JSON file located at `/home/.badger/config.json`. Here's an example:

```json
{
  "LogFileName": "badger-33920:3f93b037-924b-451c-abc1-9766e1d2b923.log",
  "PidFileName": "badger-33920:3f93b037-924b-451c-abc1-9766e1d2b923.pid",
  "ClusterName": "badger:allworkers",
  "CustomQueues": {
    "Queues": {
      "mycustomqueue": {
        "Concurrency": 5,
        "Timeout": 60,
        "DoneLog": true
      },
      "mailqueue": {
        "Concurrency": 10,
        "Timeout": 60
      },

    }
  },
  "RedisURL": "redis://localhost:6379"
}
```

### ✏️ Customize Your Queues

Update the `CustomQueues` section to define your own queues. Each queue must have:
- A unique name (e.g., "mailqueue", "alertsqueue")
- `Concurrency`: number of parallel workers for that queue
- `Timeout`: how long to wait before considering the job failed (in seconds)
- `DoneLog` (optional): set to `true` to log completed jobs

The worker processes jobs as shell commands, making it versatile for various automation tasks.

## 🧠 Usage

Run the worker normally:

```sh
./badger
```

Or run it as a background daemon:

```sh
./badger -d
```

### 🛠️ CLI Options

```sh
Usage of ./badger:
  -d    Run as a daemon in the background
```

## 🖥️ Web UI

badger comes with a built-in web interface that runs on port 5000, accessible at `http://localhost:5000`. The Web UI provides a convenient way to:

- **Monitor active jobs** - See what's currently running across all queues
- **View job history** - Check completed, failed, and delayed jobs
- **Manage queues** - Pause, resume, and inspect individual queues
- **Schedule new jobs** - Add jobs directly through the interface
- **Check worker status** - View health metrics and configuration details

![badger Web UI Dashboard](./ui_images/ui_home.png)
![badger Web UI Dashboard](./ui_images/queue.png)
![badger Web UI Dashboard](./ui_images/job_logs.png)
### Key Web UI Features

- **Queue Overview**: See at-a-glance statistics for all your configured queues
- **Job Logs**: Access detailed logs for jobs with `DoneLog: true` in their queue configuration
- **Job Control**: Retry failed jobs, cancel running jobs, or delete pending jobs
- **Search**: Find specific jobs by command, status, or queue

Access the web interface by navigating to:
```
http://localhost:5000
```

## 📝 License

[MIT License](LICENSE)
