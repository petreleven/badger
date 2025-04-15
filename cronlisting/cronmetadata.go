package cronlisting

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"worker/config"
)

const (
	MINUTE = iota
	HOUR
	DAY
	MONTH
	DAYWEEK
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
	cfg := *config.Get()
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

func isInRange(value string, t int) (bool, error) {
	if len(value) > 2 {
		divisor, err := strconv.Atoi(value[2:])
		if err != nil {
			return false, err
		}
		if t%divisor == 0 {
			return true, nil
		}
	}
	return false, nil
}

func parseCronRepeatField(field *string, level int, t time.Time) {
	*field = strings.Trim(*field, " ")
	if strings.HasPrefix(*field, "*/") {
		switch level {
		case MINUTE:
			b, err := isInRange(*field, t.Minute())
			if err != nil {
				log.Println("Error parsing Field ", *field)
			}
			if b {
				*field = fmt.Sprintf("%d", t.Minute())
				return
			}
			*field = (*field)[2:]
		case HOUR:
			b, err := isInRange(*field, t.Hour())
			if err != nil {
				log.Println("Error parsing Field ", *field)
			}
			if b {
				*field = fmt.Sprintf("%d", t.Hour())
				return
			}
			*field = (*field)[2:]
		case DAY:
			b, err := isInRange(*field, t.Day())
			if err != nil {
				log.Println("Error parsing Field ", *field)
			}
			if b {
				*field = fmt.Sprintf("%d", t.Day())
				return
			}
			*field = (*field)[2:]
		case MONTH:
			b, err := isInRange(*field, int(t.Month()))
			if err != nil {
				log.Println("Error parsing Field ", *field)
			}
			if b {
				*field = fmt.Sprintf("%d", t.Month())
				return
			}
			*field = (*field)[2:]
		case DAYWEEK:
			b, err := isInRange(*field, int(t.Weekday()))
			if err != nil {
				log.Println("Error parsing Field ", *field)
			}
			if b {
				*field = fmt.Sprintf("%d", t.Weekday())
				return
			}
			*field = (*field)[2:]
		}
		return
	}
	if *field == "*" {
		switch level {
		case MINUTE:
			*field = fmt.Sprintf("%d", t.Minute())
		case HOUR:
			*field = fmt.Sprintf("%d", t.Hour())
		case DAY:
			*field = fmt.Sprintf("%d", t.Day())
		case MONTH:
			*field = fmt.Sprintf("%d", t.Month())
		case DAYWEEK:
			*field = fmt.Sprintf("%d", t.Weekday())
		}
	}
}

func (c *Cron) GetUTC(t time.Time) (int64, error) {
	parseCronRepeatField(&c.Minute, MINUTE, t)
	parseCronRepeatField(&c.Hour, HOUR, t)
	parseCronRepeatField(&c.Day, DAY, t)
	parseCronRepeatField(&c.Month, MONTH, t)
	parseCronRepeatField(&c.DayWeek, DAYWEEK, t)
	dayofweek, err := strconv.Atoi(c.DayWeek)
	if err != nil {
		log.Println("Unable to convert weekday to int for cron:", c.Name, err)
		return -1, err
	}

	dayofweekstr := ""
	switch dayofweek {
	case 0:
		dayofweekstr = "Sunday"

	case 1:
		dayofweekstr = "Monday"

	case 2:
		dayofweekstr = "Tuesday"

	case 3:
		dayofweekstr = "Wednesday"

	case 4:
		dayofweekstr = "Thursday"

	case 5:
		dayofweekstr = "Friday"

	case 6:
		dayofweekstr = "Saturday"

	}
	layout := "2006-01-02 Monday 15:04"
	if len(c.Month) != 2 {
		c.Month = "0" + c.Month
	}
	if len(c.Day) != 2 {
		c.Day = "0" + c.Month
	}
	if len(c.Hour) != 2 {
		c.Hour = "0" + c.Hour
	}
	if len(c.Minute) != 2 {
		c.Minute = "0" + c.Minute
	}

	value := fmt.Sprintf("%d", time.Now().UTC().Year()) + "-" + c.Month + "-" + c.Day +
		" " + dayofweekstr + " " +
		c.Hour + ":" + c.Minute
	cronT, err := time.Parse(layout, value)
	if err != nil {
		log.Println("Unable to convert cron data to valid time for c:", c.Name, err)
		return -1, err
	}
	return cronT.Unix(), nil
}
