package onedrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	oauth_url = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	graph_url = "https://graph.microsoft.com/v1.0"
)

type Files struct {
	Path   string `json:"path"`
	Folder bool   `json:"folder"`
}

type Config struct {
	ClientId      string  `json:"client_id"`
	RedirectUri   string  `json:"redirect_uri"`
	ClientSecret  string  `json:"client_secret"`
	RefreshToken  string  `json:"refresh_token"`
	GrantType     string  `json:"grant_type"`
	DownloadFiles []Files `json:"downloadFiles"`
}

type token struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int64  `json:"expires_in"`
	ExtExpiresIn int64  `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Drive struct {
	*Config
	*token
	cli *http.Client
}

func NewDrive(conf *Config) (d *Drive, err error) {
	d = &Drive{
		Config: conf,
		cli:    http.DefaultClient,
	}

	err = d.setToken()
	return
}

func (d *Drive) setToken() (err error) {
	conf := d.Config

	data := url.Values{}
	data.Add("client_id", conf.ClientId)
	data.Add("redirect_uri", conf.RedirectUri)
	data.Add("client_secret", conf.ClientSecret)
	data.Add("refresh_token", conf.RefreshToken)
	data.Add("grant_type", conf.GrantType)

	payload := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, oauth_url, payload)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.cli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	t := new(token)
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		return
	}

	d.token = t

	return
}

type DownItem struct {
	Name        string
	Size        int64
	Path        string
	DownloadUrl string
}

type Collection struct {
	DownItems []DownItem
}

func (c *Collection) add(d DownItem) {
	c.DownItems = append(c.DownItems, d)
}

func newCollection(cup int) *Collection {
	d := make([]DownItem, 0, cup)

	return &Collection{d}
}

func (d *Drive) Download() (coll *Collection) {
	coll = newCollection(100)

	for _, v := range d.DownloadFiles {
		coll = d.parseFiles(v, coll)
	}

	return
}

func (d *Drive) parseChildren(path string, coll *Collection) {
	// https://learn.microsoft.com/en-us/onedrive/developer/rest-api/concepts/optional-query-parameters?view=odsp-graph-online
	url := fmt.Sprintf(
		"%s/me/drive/root:%s:/children?select=name,size,folder,file,@microsoft.graph.downloadUrl",
		graph_url,
		path,
	)

	items, err := d.parseItems(url, true)
	if err != nil {
		return
	}

	for _, v := range items.childrens.Value {
		if v.Folder.ChildCount != 0 {
			// 递归遍历目录
			subPath := path + "/" + v.Name
			d.parseChildren(subPath, coll)
		} else {
			coll.add(DownItem{
				Name:        v.Name,
				Size:        v.Size,
				Path:        path,
				DownloadUrl: v.DownloadUrl,
			})
		}
	}
}

func (d *Drive) parseFiles(f Files, c *Collection) (coll *Collection) {
	coll = c

	if f.Folder {
		d.parseChildren(f.Path, coll)
	} else {
		url := fmt.Sprintf(
			"%s/me/drive/root:%s?select=name,size,file,@microsoft.graph.downloadUrl",
			graph_url,
			f.Path,
		)

		items, err := d.parseItems(url, f.Folder)
		if err != nil {
			return
		}
		i := items.item
		coll.add(DownItem{
			Name:        i.Name,
			Size:        i.Size,
			Path:        f.Path,
			DownloadUrl: i.DownloadUrl,
		})
	}

	return
}

type item struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadUrl string `json:"@microsoft.graph.downloadUrl"`
}

type children struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadUrl string `json:"@microsoft.graph.downloadUrl"`
	Folder      struct {
		ChildCount int64 `json:"childCount"`
	} `json:"folder"`
}

type childrens struct {
	Value []children `json:"value"`
}

type items struct {
	childrens *childrens
	item      *item
}

func newItems(childrenCup int) *items {
	i := make([]children, 0, childrenCup)
	childrens := childrens{i}
	item := item{}

	return &items{
		childrens: &childrens,
		item:      &item,
	}
}

func (d *Drive) parseItems(url string, isFolder bool) (items *items, err error) {
	items = newItems(100)
	bearer := "Bearer " + d.token.AccessToken

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", bearer)

	resp, err := d.cli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if isFolder {
		err = json.NewDecoder(resp.Body).Decode(items.childrens)
	} else {
		err = json.NewDecoder(resp.Body).Decode(items.item)
	}

	return
}
