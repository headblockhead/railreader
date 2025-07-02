package main

import (
	"fmt"
	"github.com/segmentio/kafka-go"
)

func main() {
	conn, err := kafka.Dial("tcp", "pkc-z3p1v0.europe-west2.gcp.confluent.cloud:9092")
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()
}
