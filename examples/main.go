package main

import (
	"fmt"

	"github.com/m-mizutani/gelfconv"
)

func main() {
	// example1()
	example2()
}

type logData struct {
	IPAddr  string `json:"ipaddr"`
	Port    int    `json:"port"`
	Request string `json:"request"`
}

func example1() {
	log := logData{"10.1.2.3", 51234, "GET xxx"}

	msg := gelfconv.NewMessage("test message")
	msg.SetData(log)
	rawGELF, err := msg.Gelf()
	if err != nil {
		fmt.Errorf("convert error %v", err)
	}

	fmt.Println(string(rawGELF))
}

func example2() {
	data := map[string]interface{}{
		"k1": "v1",
		"k2": map[string]string{
			"k3": "v3",
		},
		"k4": []int{1, 2, 3},
	}
	m := gelfconv.NewMessage("test")
	m.SetData(data)
	rawGELF, _ := m.Gelf()
	fmt.Println(string(rawGELF))
}
