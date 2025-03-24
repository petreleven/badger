package cronlisting

import (
	"encoding/json"
	"fmt"
)

type Cron struct {
	Name string

	Minute  string
	Hour    string
	Day     string
	Month   string
	DayWeek string
	Job     string
	Queue   string
}

func (c *Cron) Encode() string {
	return c.Minute + " " + c.Hour + " " + c.Day + " " + c.Month + " " + c.DayWeek + " " + c.Job
}

func (c *Cron) DecodeFromSlice(cronName string, cronDetails []string) error {
	if len(cronDetails) < 5 {
		return fmt.Errorf("Cron details len doesnt match expected length\n")
	}

	c.Name = cronName
	c.Minute = cronDetails[0]
	c.Hour = cronDetails[1]
	c.Day = cronDetails[2]
	c.Month = cronDetails[3]
	c.DayWeek = cronDetails[4]
	c.Job = "eg run bash shell"
	c.Queue = cfg.PendingQueue

	return nil
}

func (c *Cron) Json() (data []byte) {
	data, err := json.Marshal(c.Job)
	if err != nil {
		return nil
	}
	return data
}
