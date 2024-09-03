package main

import (
	"log"
	"mediadownload/application/onedrive"
	"mediadownload/config"
	"mediadownload/download"
	"os"
)

func main() {
	err := config.Load("./config.json")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	conf := config.OneDriveConf()
	drive, err := onedrive.NewDrive(&conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	down := download.OneDrive{
		Collection: drive.Download(),
	}

	down.Aria2()
}
