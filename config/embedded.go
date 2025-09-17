package config

import (
	"embed"
)

//go:embed *.yaml
//go:embed scripts/*
var configsFS embed.FS
