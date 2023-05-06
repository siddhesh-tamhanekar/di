package lib

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func getVarName(name string) string {
	if name[0:1] == "*" {
		name = name[1:]

	}
	return strings.ToLower(name[0:1]) + name[1:]
}

func getCurrentModuleName() string {
	dir, _ := os.Getwd()
	gomodfile := dir + "/" + "go.mod"
	if _, err := os.Stat(gomodfile); errors.Is(err, os.ErrNotExist) {
		fmt.Println("go module does not exists")

	}
	gmodfile, err := os.ReadFile(gomodfile)
	if err != nil {
		log.Panic("go module file is not readable")
	}
	modline := strings.Split(string(gmodfile), "\n")[0]
	return strings.ReplaceAll(modline, "module ", "")

}

func WriteFile(fp string, b []byte) {
	fmt.Println("GENERATED CODE FOR", fp)

	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("Write error", err)
	}
	f.Write(b)
	f.Close()

}

func ContainsStr(strings []string, needle string) bool {
	for _, v := range strings {
		if v == needle {
			return true
		}
	}
	return false
}

var Debug bool

func logMsg(msg string) {
	if Debug == true {
		fmt.Println(msg)

	}
}
