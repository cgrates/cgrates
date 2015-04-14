package utils

import "testing"

func TestRedisGenProto(t *testing.T) {
	cmd := GenRedisProtocol("SET", "mykey", "Hello World!")
	if cmd != "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$12\r\nHello World!\r\n" {
		t.Error("Wrong redis protocol command: " + cmd)
	}
}
