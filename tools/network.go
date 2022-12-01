package tools

import (
	"fmt"
	"strings"
)

func ParseNetwork(bind string) (network, address string, err error) {
	if index := strings.Index(bind, "@"); index == -1 {
		err = fmt.Errorf("address must be network@unixsocket : %s", bind)
		return
	} else {
		network = bind[:index]
		address = bind[index+1:]
		return
	}
}
