package utils

type CmdLine = [][]byte

func CmdLine1(cmdName string, args ...string) CmdLine {
	var result = make([][]byte, len(args)+1)
	result[0] = []byte(cmdName)
	for i, arg := range args {
		result[i+1] = []byte(arg)
	}
	return result
}

func CmdLine2(cmdName string, args [][]byte) CmdLine {
	var result = make([][]byte, len(args)+1)
	result[0] = []byte(cmdName)
	for i, arg := range args {
		result[i+1] = arg
	}
	return result
}
