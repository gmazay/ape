package conf

import (
	//    "bufio"
	"encoding/json"
	"gopkg.in/yaml.v2"
	//"fmt"
	"flag"
	"os"
	"strings"
)

var filename *string

func init() {
	filename = flag.String("conf", "ape.conf", "config file")
	flag.Parse()
}

type Config struct {
	Listen        string
	MaxConnsPerIP int
	Infolog       string
	Errlog        string
	DocumentRoot  string
	Dsn           map[string]*Dsn
	Route         map[string]*Route
	Auth          map[string]string
}

type Dsn struct {
	Host   string
	User   string
	Pass   string
	Dbname string
	Params string
	Type   string
}

type Route struct {
	Method map[string]*Method
}

type Method struct {
	Query   string
	Params  []string
	Default []string
	Answer  []string
	Auth    string
	Handler string
}

func (config *Config) Read() (*Config, error) {
	//filename := "agawa.conf"
	//if len( os.Args ) >1 { filename = os.Args[1] }
	//filename := flag.String("conf", "ape.conf", "config file")
	//flag.Parse()

	file, err := os.Open(*filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if strings.HasSuffix(*filename, "yml") {
		decoder := yaml.NewDecoder(file)
		err = decoder.Decode(&config)
	} else {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
	}
	//config := new(Config)
	if err != nil {
		return nil, err
	}

	//fmt.Println(config.Applications[0].Db.Host)
	return config, err
}
