package main

import (
	"context"
	"github.com/phyrwork/segvdl"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	// Check arguments
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatalln("video url not given")
	}
	if len(args) < 2 {
		log.Fatalln("audio url not given")
	}
	if len(args) < 3 {
		log.Fatalln("output path not given")
	}
	// Parse source URLs
	var videourl, audiourl segvdl.Pattern
	if err := videourl.Parse(args[0]); err != nil {
		log.Fatalf("video url parse error: %v", err)
	}
	if err := audiourl.Parse(args[1]); err != nil {
		log.Fatalf("audio url parse error: %v", err)
	}
	// Open/reserve workspace files
	var err error
	var videofile, audiofile, outfile *os.File
	if videofile, err = ioutil.TempFile("", "*.m4s"); err != nil {
		log.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(videofile.Name())
	if audiofile, err = ioutil.TempFile("", "*.m4s"); err != nil {
		log.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(audiofile.Name())
	if outfile, err = os.Create(args[2]); err != nil {
		log.Fatalf("error opening output file: %v", err)
	}
	outfile.Close()
	// Get audio and video streams
	fetch := segvdl.Fetcher{
		Workers: 16,
		Client:  segvdl.NewClient(),
	}
	log.Println("[main] copying video stream")
	if err := fetch.CopyFromPattern(ctx, videofile, videourl); err != nil {
		log.Fatalf("error copying video stream: %v", err)
	}
	videofile.Close()
	log.Println("[main] copying audio stream")
	if err := fetch.CopyFromPattern(ctx, audiofile, audiourl); err != nil {
		log.Fatalf("error copying audio stream: %v", err)
	}
	audiofile.Close()
	// Mux video
	log.Println("[main] muxing video")
	if err := segvdl.Mux(outfile.Name(), videofile.Name(), audiofile.Name()); err != nil {
		log.Fatalf("mux error: %v", err)
	}
}
