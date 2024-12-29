package main

import (
	command "github.com/GrGLeo/LoadBalancer/src/cmd"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
  scheduleDelay int
  File          string
  zaplog        *zap.SugaredLogger 
)

func init() {
  // Create Logger
  logger := zap.Must(zap.NewDevelopment())
  zap.ReplaceGlobals(logger)
  zaplog = logger.Sugar()
  // Load .env
  err := godotenv.Load()
  if err != nil {
    zap.L().Sugar().Warn("Failed to read .env\n Env variable might not be set correctly")
  }
}

func main() {
  command.Execute()
}
