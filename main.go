package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"./parse"
	"./vm"
)

func main() {
	if len(os.Args) != 2 {
		panic("参数不正确")
	}

	env := vm.NewEnv()

	// 定义默认函数
	loadBuildins(env)


	source := os.Args[1]
	input, err := ioutil.ReadFile(source)
	if err != nil {
		panic("文件不存在")
	}

	code := string(input)

	t, err := parse.Parse(code)
	code = ""

	if err == nil {
		_, err = vm.Run(t.Root, env)
	}

	if err != nil {

		if e, ok := err.(*vm.Error); ok {
			fmt.Fprintf(os.Stderr, "%s:第%d行:第%d列: %s\n", source, e.Pos.Line, e.Pos.Column, err)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func loadBuildins(env *vm.Env) {
	env.Define("print", fmt.Print)
}
