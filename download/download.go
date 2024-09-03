package download

import (
	"fmt"
	"log"
	"mediadownload/application/onedrive"
	"mediadownload/config"
	"mediadownload/download/aria2"
)

type Downloader interface {
	Aria2() error
}

type OneDrive struct {
	Collection *onedrive.Collection
}

func (o *OneDrive) Aria2() error {
	id := 1000
	conf := config.Aria2Conf()
	opts := aria2.NewOptions(&conf)

	for _, v := range o.Collection.DownItems {
		uri := []string{v.DownloadUrl}
		opts.SetUri(uri)
		opts.SetDir(v.Path)
		opts.SetId(id)

		resp, err := aria2.Download(opts)
		if err != nil {
			log.Println(err)
			continue
		}

		if resp.Error != nil {
			log.Println(resp.Error.Message)
			continue
		}

		fmt.Printf("[INFO] id: %d, name: %s, size: %d\n", resp.Id, v.Name, v.Size)

		id++
	}

	return nil
}
