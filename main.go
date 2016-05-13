package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/robertkrimen/otto/parser"
)

// type knobRecord map[string]interface{}
type knobRecord struct {
	Domain       string `json:"domain"`
	DictType     string `json:"dict_type,omitempty"`
	Key          string `json:"key,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
}

type knob struct {
	Identifiers []string     `json:"identifiers"`
	Name        string       `json:"knob_name"`
	Data        []knobRecord `json:"knob_fields,omitempty"`
}

func main() {
	var newKnobs []knob
	knobs := parseKnobs()
	ids := parseIdentifiers()
	for _, id := range ids {
		for _, k := range knobs {
			if id.knobset == k.Name {
				k.Identifiers = append(k.Identifiers, id.id)
				newKnobs = append(newKnobs, k)
			}
		}
	}

	err := json.NewEncoder(os.Stdout).Encode(newKnobs)
	if err != nil {
		log.Fatal(err)
	}

}

type identifier struct {
	knobset string
	id      string
}

func parseIdentifiers() []identifier {
	path := "/Applications/Server.app/Contents/ServerRoot/usr/share/devicemgr/frontend/admin/common/app/javascript-packed.js"
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	prsr := parser.NewParser(path, string(file))
LOOP:
	for {
		// var possibleID string
		_, literal, _ := prsr.Scan()
		switch literal {
		case "loadInitialData":
			break LOOP
		case "switch":
			tk := next(prsr)
			if tk != "s" {
				continue LOOP
			}
			tk = next(prsr)
			if tk != "PayloadType" {
				continue LOOP
			}
			return readSwitch(prsr)
		}
	}
	return nil
}

func readSwitch(prsr parser.Parser) []identifier {
	var ids []identifier
LOOP:
	for {
		tk := next(prsr)
		if tk == "default" {
			break LOOP
		}
		if tk == "case" {
			var id string
			i := next(prsr)
			newID, err := strconv.Unquote(i)
			if err != nil {
				log.Fatal(err)
				id = i
			} else {
				id = newID
			}
		INNER:
			for {
				tk := next(prsr)
				if tk == "break" {
					continue LOOP
				}
				if tk == "Admin" {
					knobset := next(prsr)
					ids = append(ids, identifier{knobset, id})
					break INNER
				}
			}

		}
	}
	return ids

}

func parseKnobs() []knob {
	path := "/Applications/Server.app/Contents/ServerRoot/usr/share/devicemgr/frontend/admin/common/app/javascript-packed.js"
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var knobs []knob
	prsr := parser.NewParser(path, string(file))
LOOP:
	for {
		var possibleKnob string
		_, literal, _ := prsr.Scan()
		switch literal {
		case "loadInitialData":
			break LOOP
		case "Admin":
			var currentKnob knob
			tk := next(prsr)
			if tk == "loadInitialData" {
				break LOOP
			}
			if !strings.HasSuffix(tk, "KnobSet") {
				continue LOOP
			}
			possibleKnob = tk
			_ = possibleKnob

			tk = next(prsr)
			if tk == "Admin" {
				currentKnob.Name = possibleKnob
			}

			tk = next(prsr)
			if tk != "KnobSet" {
				continue LOOP
			}
			tk = next(prsr)
			if tk != "extend" {
				continue LOOP
			}
			for {

				c := readRecord(prsr, &currentKnob)
				if !c {
					break
				}
			}
			knobs = append(knobs, currentKnob)

		}
	}
	return knobs
}

func readRecord(prsr parser.Parser, currentKnob *knob) bool {
	var currentRecord knobRecord
LOOP:
	for {
		possibleField := readField(prsr)
		switch possibleField {
		case "validatedProperties":
			return false
		case "tabValue", "init":
			continue LOOP
		}
		tk := next(prsr)
		if tk == "function" {
			currentRecord.DictType = "function"
			currentRecord.Domain = possibleField
			currentKnob.Data = append(currentKnob.Data, currentRecord)
			return true
		}
		if tk != "SC" {
			continue LOOP
		}
		tk = next(prsr)
		if tk != "Record" {
			continue LOOP
		}
		tk = next(prsr)
		if tk != "attr" {
			continue LOOP
		}
		dictType := next(prsr)
		currentRecord.DictType = dictType
		currentRecord.Domain = possibleField
		readRecordFields(prsr, &currentRecord)
		currentKnob.Data = append(currentKnob.Data, currentRecord)
		return true
	}
}

func readRecordFields(prsr parser.Parser, record *knobRecord) {
LOOP:
	for {
		tkn, literal, _ := prsr.Scan()
		if literal == "function" {
			continue LOOP
		}
		if tkn.String() == "}" {
			break LOOP
		}

		switch literal {
		case "key":
			k := next(prsr)
			key, err := strconv.Unquote(k)
			if err != nil {
				record.Key = k
			} else {
				record.Key = key

			}
		case "defaultValue":
			record.DefaultValue = next(prsr)
		}
	}
}

func readField(prsr parser.Parser) string {
	var field string
LOOP:
	for {
		tkn, literal, _ := prsr.Scan()
		if tkn.String() == ":" {
			break LOOP
		}
		field = literal
	}
	return field
}
func jumpTo(prsr parser.Parser, el string) {
LOOP:
	for {
		tk := next(prsr)
		if tk == el {
			for {
				tkn, _, _ := prsr.Scan()
				if tkn.String() == "{" {
					break LOOP
				}
			}
		}
	}
}

func next(prsr parser.Parser) string {
	for {
		_, literal, _ := prsr.Scan()
		if literal != "" {
			return literal
		}
	}
}
