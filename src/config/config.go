package config

import (
    "encoding/json"
    "fmt"
    "os"
)

var Data = make(map[string]string)
var requiredConfigs = [100]string{"base", "indexFile", "404File", "mainFile", "pagesPath", "resourcesPath", "title", "buildPath", "verbose"}

func Init() {
	// first lets check if there is a parseable config file
	handleConfigFile()
}

func GetValue(key string) string {
	val, exist := Data[key]
	if !exist {
		fmt.Printf("> Missing config %s exiting server.", key)
		os.Exit(0)
	}
	return val
}

func SetValue(key string, value string) {
	Data[key] = value
}

func handleConfigFile() {
    // first we read the json data
    data, err := os.ReadFile("config.json")
    if nil != err {
        fmt.Print("> Config file could not be found or is not readable")
        os.Exit(0)
        return
    }
	// now we parse the config contents
	// lets see if the body json is valid tho
	Conf := make(map[string]string)
	err = json.Unmarshal(data, &Conf)
	if nil != err {
		fmt.Print("> Config file content is not a valid json")
		os.Exit(0)
		return
	}

	// finally we write all given configs into our config Data map ### need to change this rn we can only have required configs Oo what the fuck was i tinking
	for _, name := range requiredConfigs {
		value, ok := Conf[name]
		if ok {
			Data[name] = value
		}
	}
}
