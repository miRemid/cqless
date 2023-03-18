package provider

import (
	"fmt"
	"os"
	"path"
)

func getNamespaceSecretMountPath(userSecretPath string, namespace string) string {
	return path.Join(userSecretPath, namespace)
}
func validateSecrets(secretMountPath string, secrets []string) error {
	for _, secret := range secrets {
		if _, err := os.Stat(path.Join(secretMountPath, secret)); err != nil {
			return fmt.Errorf("unable to find secret: %s", secret)
		}
	}
	return nil
}
