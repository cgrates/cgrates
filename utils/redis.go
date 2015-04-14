package utils

import "fmt"

func GenRedisProtocol(cmd ...string) string {
	proto := ""
	proto += fmt.Sprintf("*%d\r\n", len(cmd))
	for _, arg := range cmd {
		proto += fmt.Sprintf("$%d\r\n", len(arg))
		proto += fmt.Sprintf("%s\r\n", arg)
	}
	return proto
}
