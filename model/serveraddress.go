package model

// ServerAddress in a Service Instance
type ServerAddress struct {
	Name   string `json:"name" bson:"name"`
	IP     string `json:"ip" bson:"ip"`
	Port   int    `json:"port" bson:"port"`
	Backup bool   `json:"backup" bson:"backup"`
	Type   string `json:"type" bson:"type"`
}
