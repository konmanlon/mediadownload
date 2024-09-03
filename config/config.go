package config

import (
	"encoding/json"
	"mediadownload/application/onedrive"
	"mediadownload/download/aria2"
	"os"
)

type application struct {
	OneDrive onedrive.Config `json:"onedrive"`
}

type download struct {
	Aria2 aria2.Config `json:"aria2"`
}

type full struct {
	application `json:"application"`
	download    `json:"download"`
}

var config = new(full)

func OneDriveConf() onedrive.Config {
	return config.OneDrive
}

func Aria2Conf() aria2.Config {
	return config.Aria2
}

func Load(path string) (err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		return
	}

	return
}
