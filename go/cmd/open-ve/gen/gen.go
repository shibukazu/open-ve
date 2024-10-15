package gen

import (
	"fmt"
	"log"
	"os"

	"github.com/shibukazu/open-ve/go/pkg/dsl/generator"
	"github.com/spf13/cobra"
)

func NewGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen [openapi|protobuf] <file>",
		Short: "Generate Open-VE schema file",
		Long:  "Generate Open-VE schema file",
		Run:   gen,
		Args:  validateArgs,
	}
	return cmd
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("requires exactly two arguments: [openapi|protobuf] <file>")
	}

	if args[0] != "openapi" && args[0] != "protobuf" {
		return fmt.Errorf("first argument must be 'openapi' or 'protobuf'")
	}

	if _, err := os.Stat(args[1]); os.IsNotExist(err) {
		return fmt.Errorf("the file %s does not exist", args[1])
	}

	return nil
}

func gen(cmd *cobra.Command, args []string) {
	fileType := args[0]
	filePath := args[1]

	log.Printf("Generating schema for %s file: %s", fileType, filePath)

	generator.GenerateFromOpenAPI2(filePath)
}
