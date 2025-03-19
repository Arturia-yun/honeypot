package db

import (
    "context"
    "time"
    "honeypot/logServer/pkg/config"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func Connect() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GlobalConfig.MongoDB.URI))
    if err != nil {
        return err
    }

    // 测试连接
    err = client.Ping(ctx, nil)
    if err != nil {
        return err
    }

    Client = client
    DB = client.Database(config.GlobalConfig.MongoDB.Database)
    return nil
}

func Disconnect() error {
    if Client != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        return Client.Disconnect(ctx)
    }
    return nil
}