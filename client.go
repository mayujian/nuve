package nuve

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	service string
	key     string
	host    string
	timeout time.Duration
	*http.Client
}

type Config struct {
	Timeout  int64
	UseProxy bool
	Service  string
	Key      string
	Host     string
}

type message struct {
	Message string `json:"msg"`
	Code    int    `json:"code"`
}

func NewClient(conf *Config) *Client {
	var client *http.Client
	var timeout = time.Duration(conf.Timeout) * time.Millisecond
	if timeout > 0 {
		if conf.UseProxy {
			client = &http.Client{Transport: &http.Transport{
				Dial:  (&net.Dialer{Timeout: timeout}).Dial,
				Proxy: http.ProxyFromEnvironment,
			}}
		} else {
			client = &http.Client{Transport: &http.Transport{
				Dial: (&net.Dialer{Timeout: timeout}).Dial,
			}}
		}
	} else {
		client = &http.Client{}
	}
	return &Client{
		service: conf.Service,
		key:     conf.Key,
		host:    conf.Host,
		timeout: timeout,
		Client:  client,
	}
}

func calculateSignature(toSign, key []byte) string {
	mac := hmac.New(sha1.New, key)
	mac.Write(toSign)
	data := []byte(fmt.Sprintf("%x", mac.Sum(nil)))
	r := base64.StdEncoding.EncodeToString(data)
	return r
}

func (id *Client) makeHeader(user, role []byte) string {
	ts := strconv.Itoa(int(time.Now().UnixNano() / 1000000))
	cn := strconv.Itoa(int(rand.Float32() * 99999))
	sig := bytes.NewBufferString(ts)
	sig.WriteByte(',')
	sig.WriteString(cn)

	var buffer bytes.Buffer
	buffer.WriteString("MAuth realm=http://marte3.dit.upm.es,mauth_signature_method=HMAC_SHA1")
	if user != nil && role != nil {
		buffer.WriteString(",mauth_username=")
		buffer.Write(user)
		buffer.WriteString(",mauth_role=")
		buffer.Write(role)
		sig.WriteByte(',')
		sig.Write(user)
		sig.WriteByte(',')
		sig.Write(role)
	}
	signed := calculateSignature(sig.Bytes(), []byte(id.key))
	buffer.WriteString(",mauth_serviceid=")
	buffer.WriteString(id.service)
	buffer.WriteString(",mauth_cnonce=")
	buffer.WriteString(cn)
	buffer.WriteString(",mauth_timestamp=")
	buffer.WriteString(ts)
	buffer.WriteString(",mauth_signature=")
	buffer.WriteString(signed)
	return buffer.String()
}

func (id *Client) send(method, uri string, rd io.Reader, user, role []byte) ([]byte, error) {
	req, err := http.NewRequest(method, id.host+uri, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", id.makeHeader(user, role))
	req.Header.Add("Content-Type", "application/json")
	timeout := false
	timer := time.AfterFunc(id.timeout, func() {
		id.Client.Transport.(*http.Transport).CancelRequest(req)
		timeout = true
	})
	resp, err := id.Do(req)
	if timer != nil {
		timer.Stop()
	}
	if timeout {
		err = errors.New("Request time out.")
	}
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (id *Client) CreateRoom(name, options string) ([]byte, error) {
	m := make(map[string]interface{})
	m["name"] = name
	var opt = Room{Name: name}
	json.Unmarshal([]byte(options), opt)
	m["options"] = opt
	body, err := json.Marshal(&m)
	if err != nil {
		return nil, err
	}
	rd := bytes.NewReader(body)
	return id.send("POST", "/rooms/", rd, nil, nil)
}

func (id *Client) GetRooms() ([]byte, error) {
	return id.send("GET", "/rooms/", nil, nil, nil)
}

func (id *Client) GetRoom(room string) ([]byte, error) {
	return id.send("GET", "/rooms/"+room, nil, nil, nil)
}

func (id *Client) DeleteRoom(room string) ([]byte, error) {
	return id.send("DELETE", "/rooms/"+room, nil, nil, nil)
}

func (id *Client) CreateToken(room, user, role string) ([]byte, error) {
	rd := bytes.NewReader([]byte(`{}`))
	b1 := make([]byte, base64.StdEncoding.EncodedLen(len(user)))
	b2 := make([]byte, base64.StdEncoding.EncodedLen(len(role)))
	base64.StdEncoding.Encode(b1, []byte(user))
	base64.StdEncoding.Encode(b2, []byte(role))
	res, err := id.send("POST", "/rooms/"+room+"/tokens", rd, b1, b2)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (id *Client) CreateService(name, key string) ([]byte, error) {
	m := make(map[string]string)
	m["name"] = name
	m["key"] = key
	body, err := json.Marshal(&m)
	if err != nil {
		return nil, err
	}
	rd := bytes.NewReader(body)
	return id.send("POST", "/services/", rd, nil, nil)
}

func (id *Client) GetServices() ([]byte, error) {
	return id.send("GET", "/services/", nil, nil, nil)
}

func (id *Client) GetService(service string) ([]byte, error) {
	return id.send("GET", "/services/"+service, nil, nil, nil)
}

func (id *Client) DeleteService(service string, forced bool) ([]byte, error) {
	var rd io.Reader
	if forced {
		rd = strings.NewReader(`{"forced":true}`)
	} else {
		rd = strings.NewReader(`{"forced":false}`)
	}
	return id.send("DELETE", "/services/"+service, rd, nil, nil)
}

func (id *Client) GetUsers(room string) ([]byte, error) {
	return id.send("GET", "/rooms/"+room+"/users/", nil, nil, nil)
}

func (id *Client) GetUser(room, user string) ([]byte, error) {
	return id.send("GET", "/rooms/"+room+"/users/"+user, nil, nil, nil)
}

func (id *Client) DeleteUser(room, user string) ([]byte, error) {
	return id.send("DELETE", "/rooms/"+room+"/users/"+user, nil, nil, nil)
}
