package gen

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/shibukazu/open-ve/go/pkg/dsl/generator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen [openapi|protobuf] <api schema file> <output dir>",
		Short: "Generate Open-VE schema file",
		Long:  "Generate Open-VE schema file",
		Run:   gen,
		Args:  validateArgs,
	}
	return cmd
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("requires exactly three arguments: [openapi|protobuf] <api schema file> <output dir>")
	}

	if args[0] != "openapi" && args[0] != "protobuf" {
		return fmt.Errorf("first argument must be 'openapi' or 'protobuf'")
	}

	if _, err := os.Stat(args[1]); os.IsNotExist(err) {
		return fmt.Errorf("the api schema file %s does not exist", args[1])
	}

	return nil
}

func gen(cmd *cobra.Command, args []string) {
	fileType := args[0]
	filePath := args[1]
	outputDir := args[2]

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("🏭 generating open-ve schema", slog.String("fileType", fileType), slog.String("filePath", filePath), slog.String("outputDir", outputDir))

	var serialized []byte
	if fileType == "openapi" {
		dsl, err := generator.GenerateFromOpenAPI2(logger, filePath)
		if err != nil {
			panic(fmt.Errorf("failed to generate schema: %w", err))
		}
		serialized, err = yaml.Marshal(dsl)
		if err != nil {
			panic(fmt.Errorf("failed to serialize schema: %w", err))
		}
	} else if fileType == "protobuf" {
		panic("protobuf is not supported yet")
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.yml", time.Now().Format("20060102150405")))
	// create dir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create output dir: %w", err))
	}
	// write file
	if err := os.WriteFile(outputPath, serialized, 0644); err != nil {
		panic(fmt.Errorf("failed to write file: %w", err))
	}

	logger.Info("✅ generated open-ve schema", slog.String("outputPath", outputPath))
}
