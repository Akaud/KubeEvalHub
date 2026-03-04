package main

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/Akaud/KubeEvalHub/app"
)

func main() {
	_ = godotenv.Load()
	os.Exit(app.Run())
}
