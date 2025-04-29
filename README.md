# 🦡 BadgerWorker

A powerful, distributed cron worker implementation in Go that uses Redis for job queuing and locking. Perfect for environments where multiple instances need to coordinate task execution without conflicts.

You can run multiple workers on the same host or across different hosts.

Inspired by [brooce](https://github.com/SergeyTsalkov/brooce).

## 📋 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
  - [Basic Usage](#basic-usage)
  - [Adding Custom Queues](#adding-custom-queues)
  - [Running Multiple Workers](#running-multiple-workers)
  - [CLI Options](#cli-options)
- [Web UI](#-web-ui)
  - [Key Web UI Features](#key-web-ui-features)
- [License](#-license)

## 🚀 Features

- **Simple config-driven setup** – Minimal CLI flags, just a config file and optional daemon mode.
- **Distributed concurrency** – Multiple workers handle jobs cooperatively.
- **Redis-powered** – Uses Redis for queueing and distributed locking.
- **Daemon mode** – Run the worker in the background with the `-d` flag.
- **Web UI** – Monitor and manage jobs through a browser-based interface.

## 📦 Installation

Clone the repository and build the binary:

```sh
git clone https://github.com/petreleven/badger.git
cd badger
go build -o badger badger.go
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
      }
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

### Basic Usage

1. Make sure Redis is running and accessible at the URL specified in your config file.

2. Run the worker:

```sh
./badger
```

Or run it as a background daemon:

```sh
./badger -d
```

3. The worker will start processing jobs from the queues defined in your config file.

### Adding Custom Queues

1. Stop any running workers (if you're modifying an existing setup).

2. Edit the config file at `/home/.badger/config.json`:

```json
{
  "CustomQueues": {
    "Queues": {
      "your-new-queue": {
        "Concurrency": 3,
        "Timeout": 120,
        "DoneLog": true
      }
    }
  }
}
```

3. Save the config file and restart your worker:

```sh
./badger
```

4. Your new queue is now ready to receive and process jobs!

### Running Multiple Workers

You can run multiple worker instances to handle more jobs in parallel:

**Same Host:**

Run multiple instances with different process IDs:

```sh
# Terminal 1
./badger

# Terminal 2
./badger

# Terminal 3 - run as daemon
./badger -d
```

**Different Hosts:**

1. Make sure all hosts can access the same Redis server.
2. Configure each worker to use the same Redis URL.
3. Run the worker on each host:

```sh
./badger
```

The workers will automatically coordinate through Redis, distributing jobs across all running instances.

### CLI Options

```sh
Usage of ./badger:
  -d    Run as a daemon in the background
```

## 🖥️ Web UI

BadgerWorker comes with a built-in web interface that runs on port 5000, accessible at `http://localhost:5000`. The Web UI provides a convenient way to:

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
