package main

import (
	"fmt"

	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Println("Build version: " + buildVersion)
	fmt.Println("Build date: " + buildDate)
	fmt.Println("Build commit: " + buildCommit)

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	agentHandler := NewAgent()
	agentHandler.Run()
}
