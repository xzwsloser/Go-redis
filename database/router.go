package database

import "strings"

var commandTable map[string]*command = make(map[string]*command)

// command is the wrap of the maps between of command name of callback function
type command struct {
	name    string
	exector ExecFunc
	arity   int
}

// RegisterCommand register the maps between command and exector to commandTable
func RegisterCommand(name string, exector ExecFunc, arity int) {
	name = strings.ToLower(name)
	c := &command{
		name:    name,
		exector: exector,
		arity:   arity,
	}
	commandTable[name] = c
}
