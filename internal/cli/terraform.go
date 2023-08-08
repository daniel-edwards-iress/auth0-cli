package cli

import (
	"context"
	"errors"
	"os"
	"path"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/auth0/auth0-cli/internal/auth0"
)

var tfFlags = terraformFlags{
	OutputDIR: Flag{
		Name:      "Output Dir",
		LongForm:  "output-dir",
		ShortForm: "o",
		Help: "Output directory for the generated Terraform config files. If not provided, the files will be " +
			"saved in the current working directory.",
	},
}

type (
	terraformFlags struct {
		OutputDIR Flag
	}

	terraformInputs struct {
		OutputDIR string
	}
)

func (i *terraformInputs) parseResourceFetchers(api *auth0.API) []resourceDataFetcher {
	// Hard coding this for now until we add support for the `--resources` flag.
	return []resourceDataFetcher{
		&clientResourceFetcher{
			api: api,
		},
	}
}

func terraformCmd(cli *cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "terraform",
		Aliases: []string{"tf"},
		Short:   "Manage terraform configuration for your Auth0 Tenant",
		Long: "This command facilitates the integration of Auth0 with [Terraform](https://www.terraform.io/), an " +
			"Infrastructure as Code tool.",
	}

	cmd.SetUsageTemplate(resourceUsageTemplate())
	cmd.AddCommand(generateTerraformCmd(cli))

	return cmd
}

func generateTerraformCmd(cli *cli) *cobra.Command {
	var inputs terraformInputs

	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen", "export"}, // Reconsider aliases and command name before releasing.
		Short:   "Generate terraform configuration for your Auth0 Tenant",
		Long: "This command is designed to streamline the process of generating Terraform configuration files for " +
			"your Auth0 resources, serving as a bridge between the two.\n\nIt automatically scans your Auth0 Tenant " +
			"and compiles a set of Terraform configuration files based on the existing resources and configurations." +
			"\n\nThe generated Terraform files are written in HashiCorp Configuration Language (HCL).",
		RunE: generateTerraformCmdRun(cli, &inputs),
	}

	tfFlags.OutputDIR.RegisterString(cmd, &inputs.OutputDIR, "./")

	return cmd
}

func generateTerraformCmdRun(cli *cli, inputs *terraformInputs) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		data, err := fetchImportData(cmd.Context(), inputs.parseResourceFetchers(cli.api)...)
		if err != nil {
			return err
		}

		if err := generateTerraformConfigFiles(inputs.OutputDIR, data); err != nil {
			return err
		}

		cli.renderer.Infof("Terraform config files generated successfully.")
		cli.renderer.Infof(
			"Follow this " +
				"[quickstart](https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/quickstart) " +
				"to go through setting up an Auth0 application for the provider to authenticate against and manage " +
				"resources.",
		)

		return nil
	}
}

func fetchImportData(ctx context.Context, fetchers ...resourceDataFetcher) (importDataList, error) {
	var importData importDataList

	for _, fetcher := range fetchers {
		data, err := fetcher.FetchData(ctx)
		if err != nil {
			return nil, err
		}

		importData = append(importData, data...)
	}

	return importData, nil
}

func generateTerraformConfigFiles(outputDIR string, data importDataList) error {
	if len(data) == 0 {
		return errors.New("no import data available")
	}

	const readWritePermission = 0755
	if err := os.MkdirAll(outputDIR, readWritePermission); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if err := createMainFile(outputDIR); err != nil {
		return err
	}

	return createImportFile(outputDIR, data)
}

func createMainFile(outputDIR string) error {
	filePath := path.Join(outputDIR, "main.tf")

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileContent := `terraform {
  required_version = "~> 1.5.0"
  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "1.0.0-beta.1"
    }
  }
}

provider "auth0" {
  debug = true
}
`

	_, err = file.WriteString(fileContent)
	return err
}

func createImportFile(outputDIR string, data importDataList) error {
	filePath := path.Join(outputDIR, "auth0_import.tf")

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileContent := `# This file is automatically generated via the Auth0 CLI.
# It can be safely removed after the successful generation
# of Terraform resource definition files.
{{range .}}
import {
  id = "{{ .ImportID }}"
  to = {{ .ResourceName }}
}
{{end}}
`

	t, err := template.New("terraform").Parse(fileContent)
	if err != nil {
		return err
	}

	return t.Execute(file, data)
}
