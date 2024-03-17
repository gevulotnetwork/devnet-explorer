//go:build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golangci/golangci-lint/pkg/commands"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	buildTarget        = "./cmd/devnet-explorer"
	buildOutput        = "./target/bin/devnet-explorer"
	unitTestBinCover   = "./target/test-artifacts/coverage/bin/unit/"
	intTestBinCover    = "./target/test-artifacts/coverage/bin/int/"
	unitTestTxtCover   = "./target/test-artifacts/coverage/txt/unit.txt"
	intTestTxtCover    = "./target/test-artifacts/coverage/txt/integration.txt"
	mergedTestTxtCover = "./target/test-artifacts/coverage/txt/merged.txt"
)

func init() {
	os.Setenv(mg.VerboseEnv, "1")
}

type Go mg.Namespace

// Builds devnet-explorer binary
func (Go) Build() error {
	env := map[string]string{"CGO_ENABLED": "1"}
	return sh.RunWith(env, "go", "build", "-o", buildOutput, buildTarget)
}

// Runs devnet-explorer binary
func (Go) Run() error {
	mg.SerialDeps(Go.Build)
	return sh.Run(buildOutput)
}

// Runs devnet-explorer binary
func (Go) RunWithMockDB() error {
	mg.SerialDeps(Go.Build)
	return sh.RunWith(map[string]string{"MOCK_STORE": "true"}, buildOutput)
}

// Runs unit tests
func (Go) UnitTest() error {
	err := os.MkdirAll(unitTestBinCover, 0o755)
	if err != nil {
		return err
	}
	dir, err := filepath.Abs(unitTestBinCover)
	if err != nil {
		return err
	}
	err = sh.Run("go", "test", "-race", "-cover", "-covermode", "atomic", "./...", "-test.gocoverdir="+dir)
	if err != nil {
		return err
	}

	return createCoverProfile(unitTestTxtCover, unitTestBinCover)
}

// Runs integration tests
func (Go) IntegrationTest() error {
	err := sh.Run("go", "test", "-tags=integration", "-count=1", "./integrationtests")
	if err != nil {
		return err
	}

	return createCoverProfile(intTestTxtCover, intTestBinCover)
}

// Runs all tests and open coverage in browser
func (Go) TestAndCover() {
	mg.SerialDeps(
		Go.UnitTest,
		Go.IntegrationTest,
		Go.MergeCover,
		Go.ViewCoverage,
	)
}

func (Go) MergeCover() error {
	return createCoverProfile(mergedTestTxtCover, unitTestBinCover+","+intTestBinCover)
}

// Open test coverage in browser
func (Go) ViewCoverage(ctx context.Context) error {
	return sh.Run("go", "tool", "cover", "-html", mergedTestTxtCover)
}

// Runs golangci-lint
func (Go) Lint() error {
	os.Setenv("CGO_ENABLED", "1")
	oldArgs := make([]string, len(os.Args))
	copy(oldArgs, os.Args)
	os.Args = []string{"golangci-lint", "run", "--build-tags=integration", "./..."}
	defer func() { os.Args = oldArgs }()
	return commands.NewExecutor(commands.BuildInfo{}).Execute()
}

// Runs go mod tidy
func (Go) Tidy() error {
	return sh.Run("go", "mod", "tidy")
}

// Runs go mod tidy and verifies that go.mod and go.sum are in sync with the code
func (g Go) TidyAndVerify() error {
	if err := g.Tidy(); err != nil {
		return err
	}
	if err := sh.Run("git", "diff", "--exit-code", "--", "go.mod", "go.sum"); err != nil {
		return fmt.Errorf("go.mod and go.sum are not in sync. run `go mod tidy` and commit changes")
	}
	return nil
}

// Runs go mod tidy and verifies that go.mod and go.sum are in sync with the code
func Clean() error {
	return sh.Rm("./target")
}

func createCoverProfile(output string, inputDir string) error {
	err := os.MkdirAll(filepath.Dir(output), 0o755)
	if err != nil {
		return err
	}
	return sh.Run("go", "tool", "covdata", "textfmt", "-i="+inputDir, "-o", output)
}
