package parser

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/KitlerUA/xlsxparser/config"
	"github.com/KitlerUA/xlsxparser/policy"
	"github.com/KitlerUA/xlsxparser/xlsxparser"
)

//Parse - read and parse file with fileName
//write results to dir
//also returns warning message
func Parse(fileName, dir string) (string, error) {
	var warn string
	var err error
	//if dir empty - save on current directory
	if dir, err = checkDir(dir); err != nil {
		return warn, err
	}
	//initialize config
	if e := config.Init(); e != nil {
		return warn, e
	}
	var (
		records  map[string][][]string
		bindings map[string][][]string
		ext      string
	)

	ext = path.Ext(fileName)
	//find extension of file
	switch ext {
	case ".xlsx":
		records, bindings, warn, err = xlsxparser.XLSX(fileName)
		if err != nil {
			return warn, fmt.Errorf("cannot parse xlsx: %s", err)
		}
	default:
		return warn, fmt.Errorf("format of file isn`t supported")
	}
	//iterate throw records
	for k := range records {
		dirName := dir + time.Now().Format("2006-01-02_15-04-05") + "_" + k
		if err = os.Mkdir(dirName, os.ModePerm); err != nil && !os.IsExist(err) {
			return warn, fmt.Errorf("cannot create directory for policies: %s", err)
		}
		policies, warnings := xlsxparser.Parse(records[k], bindings[k])
		for _, w := range warnings {
			warn += fmt.Sprintf("<b>%s</b>: %s<br>", k, w)
		}
		if err = writePolicies(policies, dirName); err != nil {
			return warn, err
		}
	}
	return warn, nil
}

//ReplaceRuneWith - return copy of string with changed rune1 to rune2
func ReplaceRuneWith(str string, char1, char2 rune) string {
	buffer := bytes.Buffer{}
	for _, c := range str {
		if c == char1 {
			buffer.WriteRune(char2)
		} else {
			buffer.WriteRune(c)
		}
	}
	return buffer.String()
}

func writePolicies(policies []policy.Policy, dirName string) error {
	for _, p := range policies {
		marshaledPolicies, err := json.Marshal(&p)
		if err != nil {
			return fmt.Errorf("cannot marshal policy '%s' : %s", p.Name, err)
		}
		newName := ReplaceRuneWith(p.FileName, ':', '_')
		newName = ReplaceRuneWith(newName, '*', '_')
		if err = ioutil.WriteFile(dirName+"/"+newName+".json", marshaledPolicies, 0666); err != nil {
			return fmt.Errorf("cannot save json file for policy '%s': %s", p.Name, err)
		}
	}
	return nil
}

func checkDir(dir string) (string, error) {
	if dir == "" {
		var err error
		if dir, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			return "", err
		}
		dir += "/"
	} else {
		//if directory doesn't exist - return error
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return "", err
		}
	}
	return dir, nil
}
