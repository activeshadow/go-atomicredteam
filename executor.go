package atomicredteam

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"actshad.dev/go-atomicredteam/types"
	"gopkg.in/yaml.v3"
)

func Execute(tid, name, repo string, inputs []string) (*types.AtomicTest, error) {
	fmt.Println(string(MustAsset("logo.txt")))

	fmt.Println()

	fmt.Println("***** EXECUTION PLAN IS *****")
	fmt.Println(" Technique: " + tid)
	fmt.Println(" Test:      " + name)
	fmt.Println(" Inputs:    " + strings.Join(inputs, "\n            "))
	fmt.Println(" * Use at your own risk :) *")
	fmt.Println("***** ***************** *****")

	test, err := getTest(tid, name, repo)
	if err != nil {
		return nil, err
	}

	args, err := checkArgsAndGetDefaults(test, inputs)
	if err != nil {
		return nil, err
	}

	if err := checkPlatform(test); err != nil {
		return nil, err
	}

	if test.Executor == nil {
		return nil, fmt.Errorf("test has no executor")
	}

	var found bool

	for _, e := range types.SupportedExecutors {
		if test.Executor.Name == e {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("executor %s is not supported", test.Executor.Name)
	}

	command, err := interpolateWithArgs(test.Executor, args)
	if err != nil {
		return nil, err
	}

	var results string

	switch test.Executor.Name {
	case "bash":
		results, err = executeBash(command)
	case "command_prompt":
		results, err = executeCommandPrompt(command)
	case "manual":
		results, err = executeManual(command)
	case "powershell":
		results, err = executePowerShell(command)
	case "sh":
		results, err = executeSh(command)
	}

	if err != nil {
		return nil, err
	}

	fmt.Println("\nExecutor Results:")
	fmt.Println("**************************************************")
	fmt.Println(results)
	fmt.Println("**************************************************")

	for k, v := range test.InputArugments {
		v.ExpectedValue = args[k]
		test.InputArugments[k] = v
	}

	test.Executor.ExecutedCommand = map[string]interface{}{
		"command": command,
		"results": results,
	}

	return test, nil
}

func getTest(tid, name, repo string) (*types.AtomicTest, error) {
	orgBranch := strings.Split(repo, "/")

	if len(orgBranch) != 2 {
		return nil, fmt.Errorf("repo must be in format <org>/<branch>")
	}

	fmt.Printf("\nGetting Atomic Tests technique %s from GitHub repo %s\n", tid, repo)

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/atomic-red-team/%s/atomics/%s/%s.yaml", orgBranch[0], orgBranch[1], tid, tid)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting Atomic Test from GitHub: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Atomic Test from GitHub response: %w", err)
	}

	var plan types.Atomic

	if err := yaml.Unmarshal(body, &plan); err != nil {
		return nil, fmt.Errorf("processing Atomic Test YAML file: %w", err)
	}

	fmt.Printf("  - technique has %d tests\n", len(plan.AtomicTests))

	var test *types.AtomicTest

	for _, t := range plan.AtomicTests {
		if t.Name == name {
			test = &t
			break
		}
	}

	if test == nil {
		return nil, fmt.Errorf("could not find test %s/%s", tid, name)
	}

	fmt.Printf("  - found test named %s\n", name)

	return test, nil
}

func checkArgsAndGetDefaults(test *types.AtomicTest, inputs []string) (map[string]string, error) {
	var (
		keys    []string
		args    = make(map[string]string)
		updated = make(map[string]string)
	)

	for _, i := range inputs {
		kv := strings.Split(i, "=")

		if len(kv) == 2 {
			keys = append(keys, kv[0])
			args[kv[0]] = kv[1]
		}
	}

	fmt.Println("\nChecking arguments...")
	fmt.Println("  - supplied on command line: " + strings.Join(keys, ", "))

	for k, v := range test.InputArugments {
		fmt.Println("  - checking for argument " + k)

		val, ok := args[k]

		if ok {
			fmt.Println("   * OK - found argument in supplied args")
		} else {
			fmt.Println("   * XX not found, trying default arg")

			val = v.Default

			if val == "" {
				return nil, fmt.Errorf("argument [%s] is required but not set and has no default", k)
			} else {
				fmt.Println("   * OK - found argument in defaults")
			}
		}

		updated[k] = val
	}

	return updated, nil
}

func checkPlatform(test *types.AtomicTest) error {
	var platform string

	switch runtime.GOOS {
	case "linux", "freebsd", "netbsd", "openbsd", "solaris":
		platform = "linux"
	case "darwin":
		platform = "macos"
	case "windows":
		platform = "windows"
	}

	if platform == "" {
		return fmt.Errorf("unable to detect our platform")
	}

	fmt.Printf("\nChecking platform vs our platform (%s)...\n", platform)

	var found bool

	for _, p := range test.SupportedPlatforms {
		if p == platform {
			found = true
			break
		}
	}

	if found {
		fmt.Println("  - OK - our platform is supported!")
	} else {
		return fmt.Errorf("unable to run test that supports platforms %v because we are on %s", test.SupportedPlatforms, platform)
	}

	return nil
}

func interpolateWithArgs(executor *types.AtomicExecutor, args map[string]string) (string, error) {
	fmt.Println("\nInterpolating command with input arguments...")

	var interpolatee string

	if executor.Name == "manual" {
		interpolatee = executor.Steps
	} else {
		interpolatee = executor.Command
	}

	interpolated := strings.TrimSpace(interpolatee)

	for k, v := range args {
		fmt.Printf("  - interpolating [#{%s}] => [%s]\n", k, v)
		interpolated = strings.ReplaceAll(interpolated, "#{"+k+"}", v)
	}

	return interpolated, nil
}

func executeCommandPrompt(command string) (string, error) {
	fmt.Printf("\nExecuting executor=cmd command=[%s]\n", command)

	output, err := exec.Command("cmd.exe", "/c", command).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("executing command via cmd.exe: %w", err)
	}

	return string(output), nil
}

func executeSh(command string) (string, error) {
	fmt.Printf("\nExecuting executor=sh command=[%s]\n", command)

	output, err := exec.Command("sh", "-c", command).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("executing command via sh: %w", err)
	}

	return string(output), nil
}

func executeBash(command string) (string, error) {
	fmt.Printf("\nExecuting executor=bash command=[%s]\n", command)

	output, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("executing command via bash: %w", err)
	}

	return string(output), nil
}

func executePowerShell(command string) (string, error) {
	fmt.Printf("\nExecuting executor=powershell command=[%s]\n", command)

	output, err := exec.Command("powershell", "-iex", command).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("executing command via powershell: %w", err)
	}

	return string(output), nil
}

func executeManual(command string) (string, error) {
	fmt.Println("\nExecuting executor=manual command=[<see below>]")

	steps := strings.Split(command, "\n")

	fmt.Printf("\nThe following steps should be executed manually:\n\n")

	for _, step := range steps {
		fmt.Printf("    %s\n", step)
	}

	return command, nil
}
