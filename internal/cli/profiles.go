package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/report"
)

func newProfilesCommand(service Application, global *globalOptions) *cobra.Command {
	listFormatValue := string(presentationText)
	command := &cobra.Command{
		Use:   "profiles",
		Short: "List custom project profiles",
		Args:  formatArguments(cobra.NoArgs, &listFormatValue),
		RunE: func(command *cobra.Command, _ []string) error {
			format, _ := parsePresentationFormat(listFormatValue)
			options := global.reportOptions(command.OutOrStdout())
			result, err := service.ListProfiles(application.ProfileRequest{
				Project: global.project,
				Config:  global.config,
			})
			if err != nil {
				return wrapPresentationError(err, format, options)
			}

			var rendered string
			if format == presentationJSON {
				rendered, err = report.ProfileListJSON(result, options)
			} else {
				rendered, err = report.ProfileList(result, options)
			}
			if err != nil {
				return wrapPresentationError(fmt.Errorf("render custom profiles: %w", err), format, options)
			}
			_, err = fmt.Fprint(command.OutOrStdout(), rendered)
			return wrapPresentationError(err, format, options)
		},
	}
	command.Flags().StringVar(&listFormatValue, "format", string(presentationText), "output format: text or json")

	showFormatValue := string(presentationText)
	show := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one effective custom profile",
		Args:  formatArguments(cobra.ExactArgs(1), &showFormatValue),
		RunE: func(command *cobra.Command, args []string) error {
			format, _ := parsePresentationFormat(showFormatValue)
			options := global.reportOptions(command.OutOrStdout())
			result, err := service.ShowProfile(application.ProfileShowRequest{
				ProfileRequest: application.ProfileRequest{
					Project: global.project,
					Config:  global.config,
				},
				ID: args[0],
			})
			if err != nil {
				return wrapPresentationError(err, format, options)
			}

			var rendered string
			if format == presentationJSON {
				rendered, err = report.ProfileShowJSON(result, options)
			} else {
				rendered, err = report.ProfileShow(result, options)
			}
			if err != nil {
				return wrapPresentationError(fmt.Errorf("render custom profile: %w", err), format, options)
			}
			_, err = fmt.Fprint(command.OutOrStdout(), rendered)
			return wrapPresentationError(err, format, options)
		},
	}
	show.Flags().StringVar(&showFormatValue, "format", string(presentationText), "output format: text or json")
	command.AddCommand(show)

	return command
}

func formatArguments(base cobra.PositionalArgs, formatValue *string) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		if err := base(command, args); err != nil {
			return err
		}
		_, err := parsePresentationFormat(*formatValue)
		return err
	}
}
