package e2e

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-cmd/cmd"
)

type E2EConfig struct {
	// e2e settings
	SorobanCLISourceVolume    string
	SorobanCLICrateVersion    string
	SorobanJSClientNpmVersion string
	SorobanExamplesGitHash    string
	SorobanExamplesRepoURL    string
	VerboseOutput             bool

	// target network that test will use
	TargetNetworkRPCURL     string
	TargetNetworkPassPhrase string
	TargetNetworkSecretKey  string
	TargetNetworkPublicKey  string
}

const TestTmpDirectory = "test_tmp_workspace"

func InitEnvironment() (*E2EConfig, error) {
	var flagConfig = &E2EConfig{}
	var err error

	flagConfig.SorobanCLICrateVersion, _ = getEnv("SorobanCLICrateVersion")
	flagConfig.SorobanCLISourceVolume, _ = getEnv("SorobanCLISourceVolume")

	//TODO - enable this when JS Client test steps are supported.
	flagConfig.SorobanJSClientNpmVersion, _ = getEnv("SorobanJSClientNpmVersion")

	if flagConfig.SorobanExamplesGitHash, err = getEnv("SorobanExamplesGitHash"); err != nil {
		return nil, err
	}
	if flagConfig.SorobanExamplesRepoURL, err = getEnv("SorobanExamplesRepoURL"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkRPCURL, err = getEnv("TargetNetworkRPCURL"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkPassPhrase, err = getEnv("TargetNetworkPassPhrase"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkSecretKey, err = getEnv("TargetNetworkSecretKey"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkPublicKey, err = getEnv("TargetNetworkPublicKey"); err != nil {
		return nil, err
	}
	if verboseOutput, err := getEnv("VerboseOutput"); err == nil {
		flagConfig.VerboseOutput, _ = strconv.ParseBool(verboseOutput)
	}

	return flagConfig, nil
}

type TestContextKey string

var TestConfigContextKey = TestContextKey("TestConfig")

func RunCommand(testCmd *cmd.Cmd, config *E2EConfig) (int, []string, error) {
	// Run, stream output, and wait for Cmd to return Status
	if config.VerboseOutput {
		fmt.Printf("running command %s %v \n\n", testCmd.Name, testCmd.Args)
	}

	output := []string{}

	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	envCmd := cmd.NewCmdOptions(cmdOptions, testCmd.Name, testCmd.Args...)
	envCmd.Dir = testCmd.Dir

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for envCmd.Stdout != nil || envCmd.Stderr != nil {
			select {
			case line, open := <-envCmd.Stdout:
				if !open {
					envCmd.Stdout = nil
					continue
				}
				if config.VerboseOutput {
					fmt.Fprintln(os.Stdout, line)
				}
				output = append(output, line)
			case line, open := <-envCmd.Stderr:
				if !open {
					envCmd.Stderr = nil
					continue
				}
				if config.VerboseOutput {
					fmt.Fprintln(os.Stderr, line)
				}
			}
		}
	}()

	// Run and wait for Cmd to return, discard Status
	<-envCmd.Start()

	// Wait for goroutine to print everything
	<-doneChan

	return envCmd.Status().Exit, output, envCmd.Status().Error
}

func InstallCli(fc *E2EConfig) error {

	var installCliCmd *cmd.Cmd
	if fc.SorobanCLISourceVolume != "" {
		installCliCmd = cmd.NewCmd("cargo",
			"install",
			"--config", "net.git-fetch-with-cli=true",
			"--config", "build.jobs=6",
			"-f",
			"--path", "./cmd/soroban-cli")
		installCliCmd.Dir = fc.SorobanCLISourceVolume
	} else if fc.SorobanCLICrateVersion == "" {
		envCmd := cmd.NewCmd("soroban", "version")

		status, versionOutput, err := RunCommand(envCmd, fc)

		if status != 0 || err != nil {
			return fmt.Errorf("no soroban cli present, SorobanCLICrateVersion was not specified, and not able to run soroban from current path, %d, %e", status, err)
		}
		fmt.Printf("SorobanCLICrateVersion was not specified, will use version already present on path:\n %v \n\n", versionOutput)
		return nil
	} else {
		installCliCmd = cmd.NewCmd("cargo",
			"install",
			"--config", "net.git-fetch-with-cli=true",
			"--config", "build.jobs=6",
			"-f", "--locked", "soroban-cli",
			"--version", fc.SorobanCLICrateVersion)
	}

	status, _, err := RunCommand(installCliCmd, fc)

	if status != 0 || err != nil {
		return fmt.Errorf("cargo install of soroban cli had status %v and error %v", status, err)
	}

	return nil
}

// asserter is used to be able to retrieve the error reported by the called assertion
type Asserter struct {
	Err error
}

// Errorf is used by the called assertion to report an error
func (a *Asserter) Errorf(format string, args ...interface{}) {
	a.Err = fmt.Errorf(format, args...)
}

func getEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("missing required env variable %s", key)
}
