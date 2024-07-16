package main

import (
	"flag"
	"fmt"
	"path/filepath"
)

// ./ours --method=example  --root=.
// ./ours --method=generate --config=./example-shard-config.json --root=./mytestnet
// ./ours --method=start --root=./mytestnet --start-time=0:53  --wait-time="20s"

func main() {
	rootDir := flag.String("root", ".", "Root directory")
	jsonDir := flag.String("config", "./example-shard-config.json", "Shard topology json file")
	var startTimeStr, method, waitTime string
	flag.StringVar(&startTimeStr, "start-time", "", "Program start time in the format HH:MM")
	flag.StringVar(&waitTime, "wait-time", "10", "System sleep time delay, in seconds, please enter a positive integer")
	flag.StringVar(&method, "method", "start", "Commands to use")
	flag.Parse()

	fmt.Println(*rootDir)

	if method == "generate" {
		GenerateConfigFiles(*jsonDir, *rootDir)
	} else if method == "start" {
		InitNode(*rootDir, startTimeStr, waitTime)
	} else if method == "example-shard-config" {
		config := ExampleShardConfig()
		config.WriteJSONToFile(filepath.Join(*rootDir, "example-shard-config.json"))
	} else {
		panic("An undefined command")
	}
}
