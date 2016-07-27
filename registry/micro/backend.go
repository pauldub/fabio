package micro

import (
	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/registry"
	"github.com/micro/go-micro/cmd"
	mRegistry "github.com/micro/go-micro/registry"

	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

var ErrMissingRegistry = errors.New("the requested registry is not supported")

type be struct {
	reg             mRegistry.Registry
	refreshInterval time.Duration
	prefix          string
}

func NewBackend(cfg config.Micro) (registry.Backend, error) {
	refreshInterval, err := time.ParseDuration(cfg.RefreshInterval)
	if err != nil {
		return nil, err
	}

	reg, ok := cmd.DefaultRegistries[cfg.Registry]
	if !ok {
		return nil, ErrMissingRegistry
	}

	return &be{
		reg:             reg(mRegistry.Addrs(cfg.RegistryAddress)),
		refreshInterval: refreshInterval,
		prefix:          cfg.Prefix,
	}, nil
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
			svcs, err := b.reg.ListServices()
			if err != nil {
				log.Printf("[WARN] micro: failed to get services: %+v", err)

				// FIXME: Try to acquire a new watcher
				watcher.Stop()
				close(svc)
				continue
			}

			config := make([]string, 0)

			for _, sv := range svcs {
				services, err := b.reg.GetService(sv.Name)
				if err != nil {
					log.Printf("[WARN] micro: failed to get service: %+v", err)
					continue
				}

				for _, service := range services {
					for _, node := range service.Nodes {
						for _, route := range findRoutes(node.Metadata, b.prefix) {
							config = append(config, fmt.Sprintf(
								"route add %s %s%s http://%s:%d/ tags %q",
								service.Name, route.host, route.path, node.Address, node.Port,
								formatTags(service, node.Metadata, b.prefix),
							))
						}
					}
				}
			}

			svc <- strings.Join(config, "\n")

			time.Sleep(b.refreshInterval)
		}
	}()

	return svc
}

func formatTags(service *mRegistry.Service, meta map[string]string, prefix string) string {
	tags := make([]string, 0)
	tags = append(tags, fmt.Sprintf("version:%s", service.Version))

	for k, v := range meta {
		if strings.HasPrefix(k, prefix) {
			continue
		}

		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	return strings.Join(tags, ",")
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
