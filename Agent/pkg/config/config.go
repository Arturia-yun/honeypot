package config

import (
    "gopkg.in/ini.v1"
    "honeypot/Agent/pkg/vars"
)

func LoadConfig(configPath string) error {
    cfg, err := ini.Load(configPath)
    if err != nil {
        return err
    }

    section := cfg.Section("client")
    vars.GlobalConfig.Client.Interface = section.Key("INTERFACE").String()
    vars.GlobalConfig.Client.ManagerURL = section.Key("MANAGER_URL").String()
    vars.GlobalConfig.Client.Key = section.Key("KEY").String()
    vars.GlobalConfig.Client.ProxyFlag = section.Key("PROXY_FLAG").MustBool(false)

    return nil
}