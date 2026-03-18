package exit

import (
	"fmt"
	"os"
)

func WithError(err error) {
	fmt.Fprintf(os.Stderr, "\033[31m•\033[0m %v\n", err)
	os.Exit(1)
}
