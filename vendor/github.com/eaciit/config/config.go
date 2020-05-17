package config

import (
	"encoding/json"
	//"errors"
	//"fmt"
	. "github.com/creativelab/toolkit"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
)

var filename string
var isLoaded bool
var configs map[string]interface{}

func getType(t interface{}) string {
	return reflect.TypeOf(t).String()
}

func Load() error {
	var err error = nil
	if isConfigFileExist() == false {
		err := ioutil.WriteFile(filename, []byte("{}"), 0644)
		if err != nil {
			return err
		}
		configs = map[string]interface{}{}
	} else {
		fileName := configFileName()
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(data, &configs); err != nil {
			return err
		}
	}
	isLoaded = true
	return err
}

func Write() error {
	var jsonBytes []byte
	if jsonStr, err := json.MarshalIndent(&configs, "", "\t"); err != nil {
		return err
	} else {
		jsonBytes = []byte(jsonStr)
	}
	fileName := configFileName()
	if err := ioutil.WriteFile(fileName, jsonBytes, 0644); err != nil {
		return err
	}
	return nil
}

func SetConfigFile(pathtofile string) error {
	if pathtofile == "" {
		pathtofile = configFileName()
	}
	filename = pathtofile
	return Load()
}

func configFileName() string {
	if filename == "" {
		filename = filepath.Join(PathDefault(false), "config.json")
	}
	return filename
}

func isConfigFileExist() bool {
	_, err := os.Stat(configFileName())
	return os.IsNotExist(err) == false
}

func HasKey(id string) bool {
	_, exist := configs[id]
	return exist
}

func Get(id string) interface{} {
	if !isLoaded {
		Load()
	}
	ret, exist := configs[id]
	if exist == false {
		ret = ""
	}
	return ret
}

func GetDefault(id string, def interface{}) interface{} {
	if !isLoaded {
		Load()
	}
	ret, exist := configs[id]
	if exist == false {
		ret = def
	}
	return ret
}

func Set(id string, value interface{}) error {
	if configs == nil {
		configs = make(map[string]interface{})
	}
	configs[id] = value
	return nil
}
