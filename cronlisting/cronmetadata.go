package cronlisting

import "encoding/json"

type Cron struct {
	Name string

	Minute string
	Hour   string
	Day    string
	Month  string
	Job    string
	Queue  string
}

func (c *Cron) Encode() string {
	return c.Minute + " " + c.Hour + " " + c.Day + " " + c.Month + " " + c.Job
}

func (c *Cron) Json() (data []byte) {
	var d struct {
		Job string
	}
	d.Job = c.Job
	data, err := json.Marshal(d)
	if err != nil {
		return nil
	}
	return
}
