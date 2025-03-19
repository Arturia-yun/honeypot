package vars

type Config struct {
    Client struct {
        Interface   string
        ManagerURL  string
        Key         string
        ProxyFlag   bool
    }
}

var GlobalConfig Config