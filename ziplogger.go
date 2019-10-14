package ziplogger

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LoggerDetails values required for functionality to work upon.
type LoggerDetails struct {
	LogFilePath  string
	ZipFrequency int
}

type cronJobTicker struct {
	timer  *time.Timer
	ticked int
}

//IntervalPeriod Time period after which CRON Job will be scheduled
const IntervalPeriod time.Duration = 24 * time.Hour

//HourToTick Hour at which CRON Job will run
const HourToTick int = 17

//MinuteToTick Minute at which CRON Job will run
const MinuteToTick int = 9

//SecondToTick Second at which CRON Job will run
const SecondToTick int = 0

// Init Function which will schedule CRON JOB
func (logger LoggerDetails) Init() {
	go setUpCRON(logger)
}

// setUpCRON Function which will schedule CRON JOB
func setUpCRON(logger LoggerDetails) {
	jobTicker := &cronJobTicker{
		ticked: 0,
	}
	jobTicker.updateTimer()
	for {
		<-jobTicker.timer.C
		fmt.Println(time.Now(), "- just ticked")
		jobTicker.ticked = jobTicker.ticked + 1
		jobTicker.cronFunctionality(logger.LogFilePath)
		if jobTicker.ticked == logger.ZipFrequency {
			zipAndDelete(logger.LogFilePath)
		}
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
	extension := filepath.Ext(logFileName)
	destinationFileName := logFileName[0 : len(logFileName)-len(extension)]
	destinationFilePath := destinationFileName + strconv.Itoa(time.Now().Year()) + strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(time.Now().Day()) + extension
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

func zipAndDelete(logFileName string) {
	var files []string
	err := filepath.Walk(filepath.Dir(logFileName), func(path string, info os.FileInfo, err error) error {
		if path == logFileName {
			return nil
		} else if filepath.Ext(path) == ".log" {
			files = append(files, path)
		} else if filepath.Ext(path) == "" {
			if strings.Contains(path, logFileName) {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	output := "done.zip"

	if err := zipFiles(output, files); err != nil {
		panic(err)
	}
	fmt.Println("Zipped File:", output)
	for _, file := range files {
		if file != logFileName {
			deleteFile(file)
		}
	}
}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func zipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

//addFileToZip Function to add files to zip.
func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func deleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		panic(err)
	}
}
