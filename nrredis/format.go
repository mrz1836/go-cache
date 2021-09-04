package nrredis

import (
	"fmt"
	"strings"
)

func formatCommand(cmd string, args []interface{}) string {
	buf := make([]string, 0, 1+len(args))
	buf = append(buf, strings.ToLower(cmd))
	for _, a := range args {
		if str, ok := a.(string); ok {
			buf = append(buf, `"`+str+`"`)
		} else {
			buf = append(buf, fmt.Sprint(a))
		}
	}
	return strings.Join(buf, " ")
}
