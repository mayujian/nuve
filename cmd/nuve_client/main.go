package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/zyxar/nuve"
)

var config struct {
	Host     string    `json:"nuve_host"`
	Timeout  int64     `json:"timeout"`
	UseProxy bool      `json:"proxy"`
	Profiles []profile `json:"profiles"`
}

type profile struct {
	Name, Id, Key string
}

var configFile, host, myprofile string
var timeout int64
var useProxy, indent bool

func init() {
	flag.StringVar(&myprofile, "profile", "", "Specify profile")
	flag.StringVar(&host, "host", "http://localhost:3000", "nuve host")
	flag.StringVar(&configFile, "conf", "config.json", "Specify config file")
	flag.Int64Var(&timeout, "timeout", 3000, "http connection timeout")
	flag.BoolVar(&useProxy, "proxy", false, "use http proxy from environment")
	flag.BoolVar(&indent, "indent", false, "petty-print json")
}

func isDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		}
		return false, errors.New(path + " exists but is not a directory")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func usage() {
	flag.Usage()
	fmt.Println(`
  supported command list:

	createroom <NAME> <OPTION>
	getrooms
	getroom <ROOM ID>
	deleteroom <ROOM ID>
	createtoken <ROOM ID> <USER> <ROLE>
	createservice <NAME> <KEY>
	getservices
	getservice <SERVICE ID>
	deleteservice <SERVICE ID>
	getusers <ROOM ID>
	getuser <ROOM ID> <USER>
	deleteuser <ROOM ID> <USER>
`)
}

func main() {
	flag.Parse()

	if _, err := os.Stat(configFile); err != nil {
		var Home string
		if usr, err := user.Current(); err != nil {
			panic(err)
		} else {
			Home = filepath.Join(usr.HomeDir, ".nuve")
		}
		exists, err := isDirExists(Home)
		if err != nil {
			panic(err)
		}
		if !exists {
			os.Mkdir(Home, 0755)
		}
		configFile = filepath.Join(Home, configFile)
	}
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error in reading configure file: %v\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(b, &config)
	if err != nil {
		fmt.Printf("Error in parsing configure file: %v\n", err)
		os.Exit(1)
	}

	if timeout > 0 {
		config.Timeout = timeout
	}
	if host != "" {
		config.Host = host
	}

	var id, key string
	if myprofile == "" {
		if len(config.Profiles) == 0 {
			println("No profile available.")
			os.Exit(1)
		}
		id = config.Profiles[0].Id
		key = config.Profiles[0].Key
	} else {
		for _, p := range config.Profiles {
			if p.Name == myprofile {
				id = p.Id
				key = p.Key
				break
			}
		}
		if id == "" || key == "" {
			fmt.Printf("Invalid profile [%s]\n", myprofile)
			os.Exit(1)
		}
	}

	c := nuve.NewClient(&nuve.Config{
		Timeout:  config.Timeout,
		UseProxy: config.UseProxy,
		Service:  id,
		Key:      key,
		Host:     config.Host,
	})
	var r []byte

	switch flag.Arg(0) {
	case "createroom", "CreateRoom":
		r, err = c.CreateRoom(flag.Arg(1), flag.Arg(2))
	case "getrooms", "GetRooms":
		r, err = c.GetRooms()
	case "getroom", "GetRoom":
		r, err = c.GetRoom(flag.Arg(1))
	case "deleteroom", "DeleteRoom":
		r, err = c.DeleteRoom(flag.Arg(1))
	case "createtoken", "CreateToken":
		r, err = c.CreateToken(flag.Arg(1), flag.Arg(2), flag.Arg(3))
		if err == nil {
			dst := make([]byte, base64.StdEncoding.DecodedLen(len(r)))
			base64.StdEncoding.Decode(dst, r)
			fmt.Printf("%s\n", r)
			r = bytes.TrimRight(dst, "\x00") // trim base64 padding
		}
	case "createservice", "CreateService":
		r, err = c.CreateService(flag.Arg(1), flag.Arg(2))
	case "getservices", "GetServices":
		r, err = c.GetServices()
	case "getservice", "GetService":
		r, err = c.GetService(flag.Arg(1))
	case "deleteservice", "DeleteService":
		var forced bool = false
		if flag.Arg(2) == "true" {
			forced = true
		}
		r, err = c.DeleteService(flag.Arg(1), forced)
	case "getusers", "GetUsers":
		r, err = c.GetUsers(flag.Arg(1))
	case "getuser", "GetUser":
		r, err = c.GetUser(flag.Arg(1), flag.Arg(2))
	case "deleteuser", "DeleteUser":
		r, err = c.DeleteUser(flag.Arg(1), flag.Arg(2))
	case "":
		usage()
	default:
		fmt.Printf("Unkown command \"%s\"\n", flag.Arg(0))
		usage()
		return
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	if indent {
		var pretty bytes.Buffer
		err = json.Indent(&pretty, r, "", "    ")
		if err != nil {
			fmt.Println(err)
			return
		}
		r = pretty.Bytes()
	}
	fmt.Printf("%s\n", r)
}
