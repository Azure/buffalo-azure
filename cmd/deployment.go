package cmd

// DeploymentParameters enables easy marshaling of ARM Template deployment parameters.
type DeploymentParameters struct {
	Schema         string                         `json:"$schema"`
	ContentVersion string                         `json:"contentVersion"`
	Parameters     map[string]DeploymentParameter `json:"parameters"`
}

// DeploymentParameter is an individual entry in the parameter list.
type DeploymentParameter struct {
	Value interface{} `json:"value,omitempty"`
}

// NewDeploymentParameters creates a new instance of DeploymentParameters with reasonable defaults but no parameters.
func NewDeploymentParameters() *DeploymentParameters {
	return &DeploymentParameters{
		Schema:         "http://schema.management.azure.com/schemas/2015-01-01/deploymentParameters.json#",
		ContentVersion: "1.0.0.0",
		Parameters:     make(map[string]DeploymentParameter),
	}
}
