package web

import (
	"fmt"
	"github.com/dustin/randbo"
	"strings"
)

func GenID() string {
	p := make([]byte, 8)
	randbo.New().Read(p)
	return fmt.Sprintf("%x", p)
}

func Key(args ...string) string {
	return strings.Join(args, ":")
}
