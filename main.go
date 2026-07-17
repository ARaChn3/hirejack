package main

import (
	"embed"

	"hirejack/cmd"
)

//go:embed frontend/dist/*
var frontend embed.FS

func main() {
	cmd.SetFrontend(frontend)
	cmd.Execute()
}
