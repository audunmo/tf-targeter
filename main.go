package main

import (
	"fmt"
	"os"

	"github.com/audunmo/tf-targeter/cmd"
	"github.com/audunmo/tf-targeter/internal/service"
	"github.com/spf13/cobra"
)

func main() {
	s := service.New()
	RootCmd.AddCommand(cmd.NewRunCmd(s))
	Execute()
}


var RootCmd = &cobra.Command{
	Use:   "tf-targeter",
	Short: "Terraform Targeter helps you interacitvely construct an apply command with mutliple explicit targets",
	Long:  "Terraform Targeter helps you interacitvely construct an apply command with mutliple explicit targets",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			panic(err)
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
