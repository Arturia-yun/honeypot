package config


type Config struct {
    MongoDB struct {
        URI      string
        Database string
    }
    APIKey string
}

var GlobalConfig Config

func Init() {
    GlobalConfig.MongoDB.URI = "mongodb://localhost:27017"
    GlobalConfig.MongoDB.Database = "honeypot"
    GlobalConfig.APIKey = "honeypot-api-key-2024"  
}