package ziplogger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// Init values required for functionality to work upon.
type Init struct {
	LogFilePath string
}

type cronJobTicker struct {
	timer *time.Timer
}

//IntervalPeriod Time period after which CRON Job will be scheduled
const IntervalPeriod time.Duration = 24 * time.Hour

//HourToTick Hour at which CRON Job will run
const HourToTick int = 14

//MinuteToTick Minute at which CRON Job will run
const MinuteToTick int = 4

//SecondToTick Second at which CRON Job will run
const SecondToTick int = 0

// SetUpCRON Function which will schedule CRON JOB
func (logger Init) SetUpCRON() {
	jobTicker := &cronJobTicker{}
	jobTicker.updateTimer()
	for {
		<-jobTicker.timer.C
		fmt.Println(time.Now(), "- just ticked")
		jobTicker.cronFunctionality(logger.LogFilePath)
		jobTicker.updateTimer()
	}
}

// updateTimer Function which will schedule next CRON JOB
func (t *cronJobTicker) updateTimer() {
	nextTick := time.Date(time.Now().Year(), time.Now().Month(),
		time.Now().Day(), HourToTick, MinuteToTick, SecondToTick, 0, time.Local)
	if !nextTick.After(time.Now()) {
		nextTick = nextTick.Add(IntervalPeriod)
	}
	fmt.Println(nextTick, "- next tick")
	diff := nextTick.Sub(time.Now())
	if t.timer == nil {
		t.timer = time.NewTimer(diff)
	} else {
		t.timer.Reset(diff)
	}
}

// cronFunctionality Function which will rename the log file and create a new log file.
func (t *cronJobTicker) cronFunctionality(logFileName string) {
	destinationFilePath := logFileName + strconv.Itoa(time.Now().Year()) + strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(time.Now().Day())
	in, err := os.Open(logFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	out, err := os.Create(destinationFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = ioutil.WriteFile(logFileName, []byte(""), 0666)
	if err != nil {
		log.Fatal(err)
	}
}
