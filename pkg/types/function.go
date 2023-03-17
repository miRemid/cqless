// https://github.com/open-faas/faas-provider/types

package types

import (
	"fmt"
)

type FunctionDeployRequest struct {
	// Service is the name of the function deployment
	Service string `json:"service"`

	// Image is a fully-qualified container image
	Image string `json:"image"`

	// Namespace for the function, if supported by the faas-provider
	Namespace string `json:"namespace,omitempty"`

	// EnvProcess overrides the fprocess environment variable and can be used
	// with the watchdog
	EnvProcess string `json:"envProcess,omitempty"`

	// EnvVars can be provided to set environment variables for the function runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Constraints are specific to the faas-provider.
	Constraints []string `json:"constraints,omitempty"`

	// Secrets list of secrets to be made available to function
	Secrets []string `json:"secrets,omitempty"`

	// Labels are metadata for functions which may be used by the
	// faas-provider or the gateway
	Labels *map[string]string `json:"labels,omitempty"`

	// Annotations are metadata for functions which may be used by the
	// faas-provider or the gateway
	Annotations *map[string]string `json:"annotations,omitempty"`

	// Limits for function
	Limits *FunctionResources `json:"limits,omitempty"`

	// Requests of resources requested by function
	Requests *FunctionResources `json:"requests,omitempty"`

	// ReadOnlyRootFilesystem removes write-access from the root filesystem
	// mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`
}

func (request FunctionDeployRequest) BuildLabels() (map[string]string, error) {
	// Adapted from faas-swarm/handlers/deploy.go:buildLabels
	labels := map[string]string{}

	if request.Labels != nil {
		for k, v := range *request.Labels {
			labels[k] = v
		}
	}

	if request.Annotations != nil {
		for k, v := range *request.Annotations {
			key := fmt.Sprintf("%s%s", DEFAULT_ANNOTATION_LABEL_PREFIX, k)
			if _, ok := labels[key]; !ok {
				labels[key] = v
			} else {
				return nil, fmt.Errorf("key %s cannot be used as a label due to a conflict with annotation prefix %s", k, DEFAULT_ANNOTATION_LABEL_PREFIX)
			}
		}
	}

	return labels, nil
}

func (request FunctionDeployRequest) BuildEnv() []string {
	envs := []string{}
	fprocessFound := false
	fprocess := "fprocess=" + request.EnvProcess
	if len(request.EnvProcess) > 0 {
		fprocessFound = true
	}

	for k, v := range request.EnvVars {
		if k == "fprocess" {
			fprocessFound = true
			fprocess = v
		} else {
			envs = append(envs, k+"="+v)
		}
	}
	if fprocessFound {
		envs = append(envs, fprocess)
	}
	return envs
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}
