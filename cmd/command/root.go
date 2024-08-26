package command

func RunMain(args []string) error {
	cmdName := ""
	if len(args) > 1 {
		cmdName = args[1]
	}

	ccmd := NewWalletCommand(cmdName)
	return ccmd.Execute()
}
