package tests

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestCloudRouter2AwsCreateConnection(t *testing.T) {
	// retryable errors in terraform testing.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../examples/fabric/v4/cloudRouterConnectivity/cloudRouter2aws",
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)
	output := terraform.Output(t, terraformOptions, "fcr2aws_connection_result")
	assert.NotNil(t, output)
}
