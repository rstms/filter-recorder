package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/poolpOrg/OpenSMTPD-framework/filter"
)

/*********************************************************************************************

 filter-recorder

 record opensmtpd filter events for testing

*********************************************************************************************/

const Version = "0.0.1"

var recordFile *os.File
var started bool

type Record map[string]any

type SessionData struct {
	dummy bool
}

func recordEvent(name string, timestamp time.Time, session filter.Session, record Record) {
	record["name"] = name
	record["timestamp"] = timestamp.UnixNano()
	record["session"] = session

	data, err := json.Marshal(&record)
	if err != nil {
		log.Fatalf("json Marshal failed: %v", err)
	}

	separator := []byte(",\n")
	if !started {
		started = true
		separator = []byte("[\n")
	}

	_, err = recordFile.Write(append(separator, data...))
	if err != nil {
		log.Fatalf("record write failed: %v", err)
	}
}

func txResetCb(timestamp time.Time, session filter.Session, messageId string) {
	recordEvent("txResetCb", timestamp, session, Record{"messageId": messageId})
}

func txBeginCb(timestamp time.Time, session filter.Session, messageId string) {
	recordEvent("txBeginCb", timestamp, session, Record{"messageId": messageId})
}

func txRcptCb(timestamp time.Time, session filter.Session, messageId string, result string, to string) {
	recordEvent("txRcptCb", timestamp, session, Record{"messageId": messageId, "result": result, "to": to})
}

func filterDataLineCb(timestamp time.Time, session filter.Session, line string) []string {
	recordEvent("filterDataLineCb", timestamp, session, Record{"line": line})
	return []string{line}
}

func cleanup(f *os.File) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	log.Println("cleanup")
	_, err := f.Write([]byte("\n]\n"))
	if err != nil {
		log.Fatalf("error writing file: %v", err)
	}
	f.Close()
	log.Println("closed")
	signal.Reset()
	os.Exit(0)
}

func main() {

	//timestamp := time.Now().Format("2006.01.02.15.04.05.0000")
	timestamp := time.Now().Format("20060102.150405.0000")
	filename := fmt.Sprintf("/tmp/%s.recording", timestamp)
	log.Printf("Starting %s v%s writing %s\n", os.Args[0], Version, filename)
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("file open failed: %v", err)
	}
	go cleanup(f)
	recordFile = f

	txBeginCb(time.Now(), filter.Session{}, "message-1")
	txBeginCb(time.Now(), filter.Session{}, "message-2")
	txBeginCb(time.Now(), filter.Session{}, "message-3")

	filter.Init()

	filter.SMTP_IN.SessionAllocator(func() filter.SessionData {
		return &SessionData{}
	})

	filter.SMTP_IN.OnTxReset(txResetCb)
	filter.SMTP_IN.OnTxBegin(txBeginCb)
	filter.SMTP_IN.OnTxRcpt(txRcptCb)
	filter.SMTP_IN.DataLineRequest(filterDataLineCb)

	filter.Dispatch()
}
