package config

import (
	"fmt"
	"github.com/liuhailove/gmiter/core/config"
	"testing"
)

func TestGetConsoleServerList(t *testing.T) {
	cfg := new(config.Entity)
	// empty
	cfg.Conf.Dashboard.Server = ""
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	// single ip
	cfg.Conf.Dashboard.Server = "112.13.223.3"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	// single domain
	cfg.Conf.Dashboard.Server = "112.13.223.3"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	// single ip including port
	cfg.Conf.Dashboard.Server = "www.dashboard.org:81"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	// mixed
	cfg.Conf.Dashboard.Server = "www.dashboard.org:81,112.13.223.3,112.13.223.4:8080,www.dashboard.org"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	// malformed
	cfg.Conf.Dashboard.Server = "www.dashboard.org:0"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	cfg.Conf.Dashboard.Server = "www.dashboard.org:-1"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	cfg.Conf.Dashboard.Server = ":80"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	cfg.Conf.Dashboard.Server = "www.dashboard.org:"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	cfg.Conf.Dashboard.Server = "www.dashboard.org:80000"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	cfg.Conf.Dashboard.Server = "www.dashboard.org:80000,www.dashboard.org:81,:80"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())

	cfg.Conf.Dashboard.Server = "https://www.dashboard.org,http://www.dashboard.org:8080,www.dashboard.org,www.dashboard.org:8080"
	config.ResetGlobalConfig(cfg)
	fmt.Println(GetConsoleServerList())
}

func TestName(t *testing.T) {
	var i interface{}
	i = 10
	var num int
	var ok bool
	if num, ok = i.(int); !ok {
		fmt.Println(ok)
	}
	fmt.Println(num)

}
