package vars

var (
    GlobalPolicyData *PolicyData
)

type Policy struct {
    Id          string   `yaml:"id"`
    WhiteIps    []string `yaml:"white_ips"`
    WhitePorts  []string `yaml:"white_ports"`
}

type BackendService struct {
    Id          string `yaml:"id"`
    ServiceName string `yaml:"service_name"`
    LocalPort   int    `yaml:"local_port"`
    BackendHost string `yaml:"backend_host"`
    BackendPort int    `yaml:"backend_port"`
}



type PolicyData struct {
    Policy  []Policy         `yaml:"policy"`
    Service []BackendService `yaml:"service"`
}