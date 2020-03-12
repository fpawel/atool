package data

import (
	"encoding/json"
	"io/ioutil"
)

func LoadFile(filename string) error {
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var x PartyValues

	if err = json.Unmarshal(jsonData, &x); err != nil {
		return err
	}
	return SetCurrentPartyValues(x)
}
