package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type CustomQueue struct {
	Priority    uint8
	Retry       uint8
	DoneLog     bool
	FailLog     bool
	Timeout     int64
	Concurrency int
}

type CustomQueues struct {
	Queues map[string]CustomQueue
}

type Config struct {
	WorkerID       string `json:"-"`
	WorkerProcName string `json:"-"`
	LogFileName    string
	PidFileName    string
	ClusterName    string
	CustomQueues   CustomQueues
	RedisURL       string
}

var (
	cfg  *Config
	lock sync.RWMutex
)

func (c *Config) SetDefaults() {
	pid := strconv.Itoa(os.Getpid())

	if c.WorkerID == "" {
		c.WorkerID = pid + ":" + uuid.NewString()
	}

	if c.WorkerProcName == "" {
		c.WorkerProcName = "badger:" + c.WorkerID
	}

	if c.LogFileName == "" {
		c.LogFileName = "badger-" + c.WorkerID + ".log"
	}

	if c.PidFileName == "" {
		c.PidFileName = "badger-" + c.WorkerID + ".pid"
	}

	if c.ClusterName == "" {
		c.ClusterName = "badger:allworkers"
	}

	if c.RedisURL == "" {
		c.RedisURL = "redis://localhost:6379"
	}

	if c.CustomQueues.Queues == nil {
		c.CustomQueues.Queues = make(map[string]CustomQueue)
	}
	for q := range c.CustomQueues.Queues {
		log.Println(q)
		if c.CustomQueues.Queues[q].Timeout == 0 {
			temp := c.CustomQueues.Queues[q]
			temp.Timeout = 60
			c.CustomQueues.Queues[q] = temp
		}
		if c.CustomQueues.Queues[q].Concurrency == 0 {
			temp := c.CustomQueues.Queues[q]
			temp.Concurrency = 1
			c.CustomQueues.Queues[q] = temp
		}

	}
}

// Save writes the current configuration to the config file
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".badger")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	configData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, configData, 0o644)
}

func Get() *Config {
	lock.RLock()
	if cfg != nil {
		defer lock.RUnlock()
		return cfg
	}
	lock.RUnlock()

	// Need to create a new config with write lock
	lock.Lock()
	defer lock.Unlock()

	// Double-check in case another goroutine created it while we were waiting
	if cfg != nil {
		return cfg
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fall back to default configuration if home directory can't be determined
		cfg = &Config{}
		cfg.SetDefaults()
		return cfg
	}

	configPath := filepath.Join(homeDir, ".badger", "config.json")

	// Check if the config file exists
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		// Create default config
		cfg = &Config{}
		cfg.SetDefaults()

		// Ensure directory exists
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0o755); err == nil {
			// Save default config
			_ = cfg.Save()
		}

		return cfg
	}

	// Read existing config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		// Fall back to default if can't read
		cfg = &Config{}
		cfg.SetDefaults()
		return cfg
	}

	// Parse the JSON into our config struct
	cfg = &Config{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		// If there's an error parsing, use defaults
		cfg = &Config{}
	}
	// Set any missing defaults
	cfg.SetDefaults()

	return cfg
}

// GetWithCustomID allows specifying a custom worker ID instead of using the PID
func GetWithCustomID(workerID string) *Config {
	config := Get()

	// If the config already has a different worker ID set, create a clone with the new ID
	if config.WorkerID != workerID {
		// Create a deep copy of the configuration
		newConfig := *config
		newConfig.WorkerID = workerID
		newConfig.WorkerProcName = "badger:" + workerID
		newConfig.LogFileName = "badger-" + workerID + ".log"
		newConfig.PidFileName = "badger-" + workerID + ".pid"

		// Don't modify the shared config
		return &newConfig
	}

	return config
}
