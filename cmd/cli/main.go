package main

import (
	"github.com/spf13/cobra"

	"github.com/crusttech/permit/internal/env"
	"github.com/crusttech/permit/internal/store/fs"
)

func main() {
	storage, err := fs.NewPermitStorage(env.GetStringEnv("STORAGE_FS_PATH", "/tmp"))
	if err != nil {
		panic(err.Error())
	}
	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(commands(storage)...)
	rootCmd.Execute()
}
