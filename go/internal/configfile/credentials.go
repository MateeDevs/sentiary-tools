package configfile

import (
	"fmt"

	"github.com/MateeDevs/sentiary-tools/go/internal/credentials"
)

// Strings are literal credentials. Only a table with exactly one string env field is a reference.
func credentialDeclaration(value any) (credentials.Project, error) {
	switch value := value.(type) {
	case nil:
		return credentials.Project{}, nil
	case string:
		return credentials.Project{InlineKey: &value}, nil
	case map[string]any:
		name, ok := value["env"].(string)
		if ok && len(value) == 1 {
			return credentials.Project{EnvironmentVariable: &name}, nil
		}
	}
	// Never include the rejected value: it may contain a literal credential.
	return credentials.Project{}, fmt.Errorf("api_key must be a literal string or a table containing only env = \"VARIABLE_NAME\"")
}
