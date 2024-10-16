package test

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/shibukazu/open-ve/go/pkg/dsl/tester"
	"github.com/shibukazu/open-ve/go/pkg/dsl/util"
	"github.com/spf13/cobra"
)

func NewTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <dsl file>",
		Short: "Test Open-VE schema file",
		Long:  "Test Open-VE schema file",
		Run:   test,
		Args:  validateArgs,
	}
	return cmd
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("requires exactly one argument: <dsl file>")
	}

	if _, err := os.Stat(args[0]); os.IsNotExist(err) {
		return fmt.Errorf("the open-ve schema file %s does not exist", args[1])
	}

	return nil
}

func test(cmd *cobra.Command, args []string) {
	filePath := args[0]

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("🧪 test open-ve schema", slog.String("filePath", filePath))

	dsl, err := util.ParseDSLYAML(filePath)
	if err != nil {
		panic(fmt.Errorf("failed to parse schema: %w", err))
	}
	result, err := tester.TestDSL(dsl)
	if err != nil {
		panic(fmt.Errorf("failed to test schema: %w", err))
	}
	for _, validationResult := range result.ValidationResults {
		if validationResult.TestCaseNotFound {
			logger.Info("❓ test not found", slog.String("id", validationResult.ID))
		} else if len(validationResult.FailedTestCases) > 0 {
			logger.Info("❌ test failed", slog.String("id", validationResult.ID), slog.String("failedTestCases", fmt.Sprintf("%v", validationResult.FailedTestCases)))
		} else {
			logger.Info("✅ test passed", slog.String("id", validationResult.ID))
		}
	}
}
