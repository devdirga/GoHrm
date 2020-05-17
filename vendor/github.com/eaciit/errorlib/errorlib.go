package errorlib

import (
	"fmt"
)

const (
	NotYetImplemented = "Method is not yet implemented"
)

func Error(packageName string, objectName string, operation string, errorMsg string) error {
	return fmt.Errorf("[%s.%s.%s] %s", packageName, objectName, operation, errorMsg)
}
