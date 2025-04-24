package database

// writeFirstKey: get the key of the write command  example: set k1 v1
func writeFirstKey(args [][]byte) ([]string, []string) {
	key := string(args[0])
	return []string{key}, nil
}

// readFirstKey: get the key of the read command example: get k1
func readFirstKey(args [][]byte) ([]string, []string) {
	key := string(args[0])
	return nil, []string{key}
}

// writeKeys: get many write keys , example: DEL k1 , k2 , k3 ...
func writeKeys(args [][]byte) ([]string, []string) {
	wks := make([]string, len(args))
	for i, arg := range args {
		wks[i] = string(arg)
	}
	return wks, nil
}

// readKeys: get many read keys , example: MGet k1 , k2 , k3 ...
func readKeys(args [][]byte) ([]string, []string) {
	rks := make([]string, len(args))
	for i, arg := range args {
		rks[i] = string(arg)
	}
	return nil, rks
}
