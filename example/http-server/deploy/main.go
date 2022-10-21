package main

import (
	"io/ioutil"

	"github.com/pulumi/pulumi-nomad/sdk/go/nomad"
	vault "github.com/pulumi/pulumi-vault/sdk/v4/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	ServiceName = "loadchecker"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		jobFile, err := ioutil.ReadFile("./job.hcl")
		if err != nil {
			return err
		}
		policyFile, err := ioutil.ReadFile("./policy.hcl")
		if err != nil {
			return err
		}

		_, err = nomad.NewJob(ctx, ServiceName, &nomad.JobArgs{
			Jobspec: pulumi.String(string(jobFile)),
			Hcl2: nomad.JobHcl2Args{
				Enabled: pulumi.BoolPtr(true),
			},
		})
		_, err = vault.NewPolicy(ctx, ServiceName, &vault.PolicyArgs{
			Name:   pulumi.String(ServiceName),
			Policy: pulumi.String(string(policyFile)),
		})

		return err
	})
}
