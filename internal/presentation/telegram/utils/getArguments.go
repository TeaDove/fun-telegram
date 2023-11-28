package utils

import "strings"

func GetArguments(message string) map[string]string {
	args := make(map[string]string, 1)

	fields := strings.Fields(message)
	if len(fields) <= 1 {
		return args
	}

	for _, field := range fields {
		arg := strings.Split(field, "=")

		if len(arg) == 2 {
			args[arg[0]] = arg[1]
		} else if len(arg) == 1 {
			args[arg[0]] = ""
		}
	}

	return args
}
