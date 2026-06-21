package main

import (
	"fmt"
	"os"
	"telegleb/internal/app"
)

func main() {
	application, err := app.InitApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "critical error during initialization: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		application.Log.Error("application runtime error", "error", err.Error())
		os.Exit(1)
	}
}