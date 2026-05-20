package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	s, err := Default()
	if err != nil {
		t.Error(err)
	}

	go s.Run()
	time.Sleep(time.Second * 2)

	t.Run("testing running server", func(t *testing.T) {
		nt, err := net.Dial("tcp", "0.0.0.0:9696")
		if err != nil {
			t.Error(err)
		}

		buf := make([]byte, 1024)
		n, err := nt.Read(buf)
		if err != nil {
			t.Error(err)
		}

		fmt.Println(string(buf[:n]))
	})
}
