package config

type Queue struct {
	UserQueue    string
	PendingQueue string
	DoneQueue    string
}

type Config struct {
	WorkerProcName      string
	LogFileName         string
	PidFileName         string
	AddPendingMasterKey string
	AddPendingLastUnix  string
	Queue
}

var cfg *Config

func (c *Config) SetDefaults() {
	c.AddPendingMasterKey = "AddPendingMaster"
	c.AddPendingLastUnix = "AddPendingLastUnix"
	c.UserQueue = "userqueue"
	c.PendingQueue = "pendinqueue"
	c.DoneQueue = "donequeue"
}

func Get() *Config {
	if cfg == nil {
		cfg = &Config{
			WorkerProcName: "badger",
			LogFileName:    "badger" + ".log",
			PidFileName:    "badger" + ".pid",
		}
		cfg.SetDefaults()

	}
	return cfg
}

