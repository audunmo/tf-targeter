package cmd

import (
	"fmt"
	"time"

	"github.com/audunmo/tf-targeter/internal/service"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

func NewRunCmd(s *service.Service) *cobra.Command {
	var userSuppliedPlanFileLocation string
	c := &cobra.Command{
		Use:   "run",
		Short: "Run tf-targeter in the current working directory",
		Long:  "Run tf-targeter in the current working directory",
		Run: func(cmd *cobra.Command, args []string) {
			spin := spinner.New(
				spinner.CharSets[4],
				100*time.Millisecond,
				spinner.WithColor("green"),
				spinner.WithFinalMSG("Plan generated\n"),
				spinner.WithSuffix(" Generating plan. This might take a while"),
			)

			if userSuppliedPlanFileLocation == "" {
				spin.Start()
				err := s.GeneratePlan()
				if err != nil {
					panic(err)
				}
				spin.Stop()

				defer func() {
					err := s.DeletePlan()
					if err != nil {
						fmt.Printf("failed to clean up tftargeter-plan. Error was: %v", err)
					}
				}()
			}

			spin = spinner.New(
				spinner.CharSets[4],
				100*time.Millisecond,
				spinner.WithColor("green"),
				spinner.WithSuffix(" Parsing plan"),
			)

			spin.Start()
			p, err := s.LoadPlan(userSuppliedPlanFileLocation)
			if err != nil {
				panic(err)
			}
			spin.Stop()

			targets := s.GetAndConfirmTargets(p)

			cmdstr := s.FormatCommand(targets)
			fmt.Printf("\n\nYour apply command is: \n\n %v", cmdstr)
		},
	}

	c.Flags().StringVarP(&userSuppliedPlanFileLocation, "planfile", "p", "", "The location of a pre-made plan, relative to the current working directory")

	return c
}
