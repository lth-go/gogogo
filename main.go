package main

import (
	"io/ioutil"
	"os"

	"./vm"
	"./parse"
)

func main() {
	if len(os.Args) != 2 {
		panic("参数不正确")
	}

	env := vm.NewEnv()

	file := os.Args[1]
	input, err := ioutil.ReadFile(file)
	if err != nil {
		panic("文件不存在")
	}

	code := string(input)

	t, err := parse.Parse(code)
	code = ""

	if err == nil {
		vm.Run(t.Root, env)
	}
}
