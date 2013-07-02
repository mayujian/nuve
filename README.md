Go impl for Licode/Nuve
======

## nuve.Client{}

``` go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/conclave/web"
	"github.com/zyxar/nuve"
	"io/ioutil"
	"log"
)

type config struct {
	ServiceId   string `json:"superserviceID"`
	ServiceKey  string `json:"superserviceKey"`
	ServiceHost string `json:"nuve_host"`
}

var nc *nuve.Client
var conf config

func init() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln("error in reading configuration")
	}
	json.Unmarshal(data, &conf)
	log.Println(conf)
	nc = nuve.NewClient(&nuve.Config{
		Service:  conf.ServiceId,
		Key:      conf.ServiceKey,
		Host:     conf.ServiceHost,
	})
}

func getServices(ctx *web.Context, val string) string {
	ctx.SetHeader("Access-Control-Allow-Origin", "*", true)
	ctx.SetHeader("Access-Control-Allow-Methods", "POST, GET, OPITIONS, DELETE", true)
	ctx.SetHeader("Access-Control-Allow-Headers", "origin, content-type", true)
	services, err := nc.GetServices()
	if err != nil {
		log.Println(err)
		return ""
	}
	return fmt.Sprintf("%v", services)
}

func main() {
	web.Get("/(.*)", getServices)
	web.Run("0.0.0.0:8086")
}
```
