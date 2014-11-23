package FlickrAPI

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)

// internals, derived/copied from
// go-flickr: A interface to Flickr's API in Google Go
// Author: Nolan Caudill
// Date: 2011-01-10
// License: MIT
// https://github.com/mncaudill/go-flickr

const (
	endpoint = "https://api.flickr.com/services/rest/?"
)

type Request struct {
	ApiKey string
	Method string
	Args   map[string]string
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func (request *Request) sign(secret string) {
	args := request.Args

	// Remove api_sig
	delete(args, "api_sig")

	sorted_keys := make([]string, len(args)+2)

	args["api_key"] = request.ApiKey
	args["method"] = request.Method

	// Sort array keys
	i := 0
	for k := range args {
		sorted_keys[i] = k
		i++
	}
	sort.Strings(sorted_keys)

	// Build out ordered key-value string prefixed by secret
	s := secret
	for _, key := range sorted_keys {
		if args[key] != "" {
			s += fmt.Sprintf("%s%s", key, args[key])
		}
	}

	// Since we're only adding two keys, it's easier
	// and more space-efficient to just delete them
	// than copy the whole map
	delete(args, "api_key")
	delete(args, "method")

	// Have the full string, now hash
	hash := md5.New()
	hash.Write([]byte(s))

	// Add api_sig as one of the args
	args["api_sig"] = fmt.Sprintf("%x", hash.Sum(nil))
}

func encodeQuery(args map[string]string) string {
	i := 0
	s := bytes.NewBuffer(nil)
	for k, v := range args {
		if i != 0 {
			s.WriteString("&")
		}
		i++
		s.WriteString(k + "=" + url.QueryEscape(v))
	}
	return s.String()
}

func (request *Request) url() string {
	args := request.Args
	args["api_key"] = request.ApiKey
	args["method"] = request.Method
	return endpoint + encodeQuery(args)
}

// Externally-visible handlers for specific API calls
// Author: Tim Burks
// Date: 2014-11-08
// License: MIT

type Connection struct {
	APIKey    string
	APISecret string
	Client    *http.Client
}

func (connection *Connection) execute(request *Request) (body []byte, ret error) {
	if connection.Client == nil {
		connection.Client = &http.Client{}
	}
	if request.ApiKey == "" || request.Method == "" {
		return nil, Error("Need both API key and method")
	}
	s := request.url()
	httpRequest, err := http.NewRequest("GET", s, nil)
	res, err := connection.Client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ = ioutil.ReadAll(res.Body)
	return body, nil
}


// flickr.photos.search

type Photo struct {
	//XMLName  xml.Name `xml:"photo"`
	Secret   string `xml:"secret,attr"`
	Farm     int    `xml:"farm,attr"`
	Title    string `xml:"title,attr"`
	IsFriend bool   `xml:"isfriend,attr"`
	IsFamily bool   `xml:"isfamily,attr"`
	Id       string `xml:"id,attr"`
	Owner    string `xml:"owner,attr"`
	Server   int    `xml:"server,attr"`
	IsPublic bool   `xml:"ispublic,attr"`
}

type Photos struct {
	//XMLName xml.Name `xml:"photos"`
	Page    int     `xml:"page,attr"`
	Pages   int     `xml:"pages,attr"`
	PerPage int     `xml:"perpage,attr"`
	Total   int     `xml:"total,attr"`
	Photos  []Photo `xml:"photo"`
}

type PhotosSearchResponse struct {
	//XMLName xml.Name `xml:"rsp"`
	Status string `xml:"stat,attr"`
	Photos Photos `xml:"photos"`
}

func (self Connection) PhotosSearch(query map[string]string, response *PhotosSearchResponse) (err error) {
	r := &Request{
		ApiKey: self.APIKey,
		Method: "flickr.photos.search",
		Args:   query,
	}
	r.sign(self.APISecret)

	body, err := self.execute(r)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(body, &response)
	return err
}

// flickr.photos.getSizes

type Size struct {
	Label  string `xml:"label,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Source string `xml:"source,attr"`
	URL    string `xml:"url,attr"`
	Media  string `xml:"media,attr"`
}

type Sizes struct {
	CanBlog     bool   `xml:"canblog,attr"`
	CanPrint    bool   `xml:"canprint,attr"`
	CanDownload bool   `xml:"candownload,attr"`
	Sizes       []Size `xml:"size"`
}

type PhotosGetSizesResponse struct {
	//XMLName xml.Name `xml:"rsp"`
	Status string `xml:"stat,attr"`
	Sizes  Sizes  `xml:"sizes"`
}

func (self Connection) PhotosGetSizes(query map[string]string, response *PhotosGetSizesResponse) (err error) {
	r := &Request{
		ApiKey: self.APIKey,
		Method: "flickr.photos.getSizes",
		Args:   query,
	}
	r.sign(self.APISecret)

	body, err := self.execute(r)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(body, &response)

	return err
}
