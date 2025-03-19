package vars

import "sync"

type ServiceConfig struct {
    Name        string `yaml:"name"`
    ListenAddr  string `yaml:"listen_addr"`
    BackendPort int    `yaml:"backend_port"`
    Credentials struct {
        Username string `yaml:"username"`
        Password string `yaml:"password"`
    } `yaml:"credentials,omitempty"`
}

type Config struct {
    API struct {
        Key       string `yaml:"key"`
        LogServer string `yaml:"log_server"`
    } `yaml:"api"`
    Proxy struct {
        Addr string `yaml:"addr"`
    } `yaml:"proxy"`
    Services map[string]ServiceConfig `yaml:"services"`
}

var (
    GlobalConfig Config
    RawIps      sync.Map // 存储攻击者IP的Map
)