package micro

import (
	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/registry"
	"github.com/micro/go-micro/cmd"
	mRegistry "github.com/micro/go-micro/registry"

	"fmt"
	"log"
	"strings"
	"time"
)

type be struct {
	reg    mRegistry.Registry
	prefix string
}

func NewBackend(cfg config.Micro) (registry.Backend, error) {
	return &be{cmd.DefaultRegistries[cfg.Registry](mRegistry.Addrs(cfg.RegistryAddress)), cfg.Prefix}, nil
}

func (b *be) Register() error {
	return nil
}

func (b *be) Deregister() error {
	return nil
}

func (b *be) ReadManual() (string, uint64, error) {
	return "", 0, nil
}

func (b *be) WriteManual(string, uint64) (bool, error) {
	return false, nil
}

func (b *be) WatchServices() chan string {
	svc := make(chan string)

	watcher, err := b.reg.Watch()
	if err != nil {
		log.Printf("[ERROR] micro: failed to get watcher: %+v", err)
		return svc
	}

	go func() {
		for {
			res, err := watcher.Next()
			if err != nil {
				log.Printf("[WARN] micro: failed to get next rseult: %+v", err)
				time.Sleep(time.Second)
				continue

			}

			service := res.Service

			switch res.Action {
			case "create":
				for _, node := range service.Nodes {
					for _, route := range findRoutes(node.Metadata, b.prefix) {
						svc <- fmt.Sprintf(
							"route add %s %s%s http://%s:%d/ tags %q",
							service.Name, route.host, route.path, node.Address, node.Port, "",
						)
					}
				}
			}

			log.Printf("[INFO] micro: res %+v", res)
		}
	}()

	return svc
}

type route struct {
	host, path string
}

func findRoutes(meta map[string]string, prefix string) []route {
	routes := make([]route, 0)

	for k, v := range meta {
		if strings.HasPrefix(k, prefix) {
			hostPath := strings.SplitN(v, "/", 2)
			host := hostPath[0]
			path := "/"

			if len(hostPath) == 2 {
				path += hostPath[1]
			}

			routes = append(routes, route{host, path})
		}
	}

	return routes
}

func (b *be) WatchManual() chan string {
	return make(chan string)
}
