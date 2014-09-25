package ttt

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
	args = append([]string{"ttt"}, args...)
	return strings.Join(args, ":")
}
