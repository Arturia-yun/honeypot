package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseLog struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Time      time.Time         `bson:"time" json:"time"`
    Protocol  string            `bson:"protocol" json:"protocol"`
    SrcIP     string            `bson:"src_ip" json:"src_ip"`
    DstIP     string            `bson:"dst_ip" json:"dst_ip"`
    Service   string            `bson:"service" json:"service"`
}

type PacketLog struct {
    BaseLog `bson:",inline"`
    SrcPort string `bson:"src_port" json:"src_port"`
    DstPort string `bson:"dst_port" json:"dst_port"`
    IsHTTP  bool   `bson:"is_http" json:"is_http"`
}

type ServiceLog struct {
    BaseLog    `bson:",inline"`
    Username   string                 `bson:"username,omitempty" json:"username,omitempty"`
    Password   string                 `bson:"password,omitempty" json:"password,omitempty"`
    Command    string                 `bson:"command,omitempty" json:"command,omitempty"`
    Data       map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
}