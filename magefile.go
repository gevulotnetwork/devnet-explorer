//go:build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golangci/golangci-lint/pkg/commands"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"github.com/olekukonko/tablewriter"
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
	os.Setenv("LOG_LEVEL", "DEBUG")
}

type (
	Go    mg.Namespace
	Git   mg.Namespace
	Image mg.Namespace
)

// Builds devnet-explorer binary
func (Go) Build() error {
	modified, err := target.Path("api/templates/index_templ.go", "api/templates/index.templ")
	if err != nil {
		return err
	}

	if modified {
		mg.SerialDeps(Go.Generate)
	}

	env := map[string]string{"CGO_ENABLED": "0"}
	return sh.RunWith(env, "go", "build", "-o", buildOutput, buildTarget)
}

// Runs devnet-explorer binary
func (Go) Run() error {
	mg.SerialDeps(Go.Build)
	return sh.Run(buildOutput)
}

// Runs devnet-explorer in mock store mode
func (Go) RunDev() error {
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

// Runs all tests and opens coverage in browser
func (Go) TestAndCover() {
	mg.SerialDeps(
		Go.UnitTest,
		Go.IntegrationTest,
		Go.MergeCover,
		Go.ViewCoverage,
	)
}

// Merges unit and integration test coverage
func (Go) MergeCover() error {
	return createCoverProfile(mergedTestTxtCover, unitTestBinCover+","+intTestBinCover)
}

// Open test coverage in browser
func (Go) ViewCoverage(ctx context.Context) error {
	return sh.Run("go", "tool", "cover", "-html", mergedTestTxtCover)
}

// Print function coverage
func (Go) FuncCoverage(ctx context.Context) error {
	mg.SerialDeps(Go.MergeCover)
	out, err := sh.Output("go", "tool", "cover", "-func", mergedTestTxtCover)
	if err != nil {
		return err
	}

	lines := strings.Split(out, "\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetHeader([]string{"Location", "Function", "Coverage"})
	for _, line := range lines {
		cols := strings.Split(line, "\t")
		cols[0] = strings.TrimPrefix(cols[0], "github.com/gevulotnetwork/devnet-explorer/")
		final := make([]string, 0, 3)
		for i := range cols {
			col := strings.TrimSpace(cols[i])
			if col != "" {
				final = append(final, col)
			}
		}
		table.Append(final)
	}
	table.Render()
	return nil
}

// Runs code generators
func (Go) Generate() error {
	return sh.Run("go", "generate", "./...")
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

// Verify that there there are no changes in the working directory
func (Git) VerifyNoChanges() error {
	return sh.Run("git", "diff", "--exit-code")
}

// Removes target directory
func Clean() error {
	return sh.Rm("./target")
}

// Builds docker image
func (Image) Build() error {
	cmd, err := dockerCmd()
	if err != nil {
		return err
	}

	args := []string{"build", "-f", "Containerfile"}
	tags := os.Getenv("TAGS")
	if tags == "" {
		tags = "dev"
	}

	for _, tag := range strings.Split(tags, " ") {
		args = append(args, "-t", "devnet-explorer:"+tag)
	}

	args = append(args, ".")
	return sh.Run(cmd, args...)
}

// Runs smoke test for image
func (Image) SmokeTest() error {
	cmd, err := dockerCmd()
	if err != nil {
		return err
	}

	tag := os.Getenv("TAG")
	if tag == "" {
		tag = "dev"
	}

	out, _ := sh.Output(cmd, "run", "--rm", "devnet-explorer:"+tag)
	if strings.Contains(out, `level=INFO msg="starting application"`) {
		fmt.Println("Smoko test passed")
		return nil
	}
	return errors.New("smoke test failed")
}

func createCoverProfile(output string, inputDir string) error {
	err := os.MkdirAll(filepath.Dir(output), 0o755)
	if err != nil {
		return err
	}
	return sh.Run("go", "tool", "covdata", "textfmt", "-i="+inputDir, "-o", output)
}

func dockerCmd() (string, error) {
	if _, err := sh.Output("podman", "version"); err == nil {
		return "podman", nil
	}

	if _, err := sh.Output("docker", "version"); err == nil {
		return "docker", nil
	}

	return "", errors.New("neither docker nor podman command available")
}
