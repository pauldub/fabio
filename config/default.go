package config

import (
	"runtime"
	"time"
)

var Default = &Config{
	ListenerValue: []string{":9999"},
	Proxy: Proxy{
		MaxConn:       10000,
		Strategy:      "rnd",
		Matcher:       "prefix",
		NoRouteStatus: 404,
		DialTimeout:   30 * time.Second,
		LocalIP:       LocalIPString(),
	},
	Registry: Registry{
		Backend: "consul",
		Consul: Consul{
			Addr:          "localhost:8500",
			Scheme:        "http",
			KVPath:        "/fabio/config",
			TagPrefix:     "urlprefix-",
			Register:      true,
			ServiceAddr:   ":9998",
			ServiceName:   "fabio",
			ServiceStatus: []string{"passing"},
			CheckInterval: time.Second,
			CheckTimeout:  3 * time.Second,
		},
		Micro: Micro{
			Registry:        "consul",
			RegistryAddress: "localhost:8500",
			Prefix:          "route-",
		},
	},
	Runtime: Runtime{
		GOGC:       800,
		GOMAXPROCS: runtime.NumCPU(),
	},
	UI: UI{
		Addr:  ":9998",
		Color: "light-green",
	},
	Metrics: Metrics{
		Prefix:   "default",
		Interval: 30 * time.Second,
	},
	CertSources: map[string]CertSource{},
}
