package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	// "github.com/faiface/beep/mp3"
	// "github.com/faiface/beep/speaker"
	// "github.com/sevlyar/go-daemon"

	alsa "github.com/cocoonlife/goalsa"
	"github.com/cryptix/wav"
	"github.com/sevlyar/go-daemon"
)

//go:embed sfx/*.wav
var sfx embed.FS

func main() {
	fmt.Println("Launching USBYamete daemon.")
	// To terminate the daemon use:
	//  kill `cat daemon.pid`
	cntxt := &daemon.Context{
		PidFileName: "daemon.pid",
		PidFilePerm: 0644,
		LogFileName: "daemon.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("Daemon started.")

	runDaemon()
}

func runDaemon() {
	fmt.Println("starting ticker")
	dc := getUSBs()
	for range time.Tick(time.Millisecond * 100) {
		switch USBcheck(&dc) {
		case 1:
			log.Print("USB inserted")
			playMP3("sfx/YameteKudasai.wav")
		case -1:
			log.Print("USB removed")
			playMP3("sfx/Moan.wav")
		}
	}
}

func USBcheck(dc *int) int {
	cc := getUSBs()
	r := 0
	if cc != *dc {
		if cc > *dc {
			r = 1
		} else if cc < *dc {
			r = -1
		}
		*dc = cc
	}
	return r
}

func linesStringCount(s string) int {
	n := strings.Count(s, "\n")
	if len(s) > 0 && !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}
func getUSBs() int {
	out, err := exec.Command("lsusb").Output()
	if err != nil {
		log.Fatal(err)
	}
	return linesStringCount(string(out))
}

func playMP3(filename string) {
	// cmd := exec.Command("aplay", file)
	// err := cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	snd, err := sfx.ReadFile(filename)
	if err != nil {
		log.Fatal(errors.New(fmt.Sprint("Embed Read:", err)))
	}

	sr := bytes.NewReader(snd)

	// wavReader
	wavReader, err := wav.NewReader(sr, int64(len(snd)))
	if err != nil {
		log.Fatal(errors.New(fmt.Sprint("WAV reader:", err)))
	}

	// require wavReader
	if wavReader == nil {
		log.Fatal(errors.New("nil wav reader"))
	}

	/*
		// print .WAV info
		// wavinfo = wavReader.String()
		fileinfo := wavReader.GetFile()
		// open default ALSA playback device
		samplerate := int(fileinfo.SampleRate)
		if samplerate == 0 {
			samplerate = 44100
		}
		if samplerate > 100000 {
			samplerate = 44100
		}
	*/
	samplerate := 44100

	out, err := alsa.NewPlaybackDevice("default", 1, alsa.FormatS16LE, samplerate, alsa.BufferParams{})
	if err != nil {
		log.Fatal(errors.New(fmt.Sprint("alsa:", err)))
	}

	// require ALSA device
	if out == nil {
		log.Fatal(errors.New("nil ALSA device"))
	}

	// close device when finished
	defer out.Close()

	for {
		s, err := wavReader.ReadSampleEvery(2, 0)
		var cvert []int16
		for _, b := range s {
			cvert = append(cvert, int16(b))
		}
		if cvert != nil {
			// play!
			out.Write(cvert)
		}
		cvert = []int16{}

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(errors.New(fmt.Sprint("WAV Decode:", err)))
		}
	}
}
