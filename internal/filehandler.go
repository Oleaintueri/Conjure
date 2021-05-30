package internal

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func ParseYaml(filename string, unmarshalTo interface{}) error {

	if _, err := os.Stat(filename); os.IsExist(err) {
		return err
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, unmarshalTo)

	if err != nil {
		return err
	}

	return nil
}
