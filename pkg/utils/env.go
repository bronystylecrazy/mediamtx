package utils

import (
	"maps"
	"os"

	"github.com/bluenviron/mediamtx/internal/externalcmd"
)

func BuildCmdEnv(paEnv map[string]string) externalcmd.Environment {
	envMap := map[string]string{}

	// 1. Start from current process env
	for _, kv := range os.Environ() {
		// split KEY=VALUE
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				envMap[kv[:i]] = kv[i+1:]
				break
			}
		}
	}

	// 2. Override/add from pa.ExternalCmdEnv()
	maps.Copy(envMap, paEnv)

	return envMap
}
