package env

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

var (
	falsy = regexp.MustCompile("(0|false|f|no|n)")
)

func GetStringEnv(key, def string) (val string) {
	var has bool

	if val, has = os.LookupEnv(key); !has {
		val = def
	}

	return
}

func GetIntEnv(key string, def int) (val int) {
	if tmp, has := os.LookupEnv(key); has {
		if i, err := strconv.ParseInt(tmp, 10, 32); err == nil {
			return int(i)
		}
	}

	return def
}

func GetBoolEnv(key string) bool {
	if val, has := os.LookupEnv(key); !has || len(val) == 0 || falsy.MatchString(strings.ToLower(val)) {
		return false
	}
	return true
}
