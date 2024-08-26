package main

import (
	"os"

	"github.com/mengbin92/wallet/cmd/command"
)

func main() {
	if err := command.RunMain(os.Args);err != nil{
		os.Exit(-1)
	}
}