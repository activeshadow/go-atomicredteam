package atomicredteam

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"actshad.dev/go-atomicredteam/types"
	"gopkg.in/yaml.v3"
)

func Execute(tid, name string, index int, inputs []string) (*types.AtomicTest, error) {
	test, err := getTest(tid, name, index)
	if err != nil {
		return nil, err
	}

	Println()

	Println("****** EXECUTION PLAN ******")
	Println(" Technique: " + tid)
	Println(" Test:      " + test.Name)

	if inputs == nil {
		Println(" Inputs:    <none>")
	} else {
		Println(" Inputs:    " + strings.Join(inputs, "\n            "))
	}

	Println(" * Use at your own risk :) *")
	Println("****************************")

	args, err := checkArgsAndGetDefaults(test, inputs)
	if err != nil {
		return nil, err
	}

	if err := checkPlatform(test); err != nil {
		return nil, err
	}

	if len(test.Dependencies) != 0 {
		var found bool

		for _, e := range types.SupportedExecutors {
			if test.DependencyExecutorName == e {
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("dependency executor %s is not supported", test.DependencyExecutorName)
		}

		Printf("\nChecking dependencies...\n")

		for _, dep := range test.Dependencies {
			Printf("  - %s", dep.Description)

			command, err := interpolateWithArgs(dep.PrereqCommand, args, true)
			if err != nil {
				return nil, err
			}

			switch test.DependencyExecutorName {
			case "bash":
				_, err = executeBash(command)
			case "command_prompt":
				_, err = executeCommandPrompt(command)
			case "manual":
				_, err = executeManual(command)
			case "powershell":
				_, err = executePowerShell(command)
			case "sh":
				_, err = executeSh(command)
			}

			if err == nil {
				Printf("   * OK - dependency check succeeded!\n")
				continue
			}

			command, err = interpolateWithArgs(dep.GetPrereqCommand, args, true)
			if err != nil {
				return nil, err
			}

			var result string

			switch test.DependencyExecutorName {
			case "bash":
				result, err = executeBash(command)
			case "command_prompt":
				result, err = executeCommandPrompt(command)
			case "manual":
				result, err = executeManual(command)
			case "powershell":
				result, err = executePowerShell(command)
			case "sh":
				result, err = executeSh(command)
			}

			if err != nil {
				if result == "" {
					result = "no details provided"
				}

				Printf("   * XX - dependency check failed: %s\n", result)

				return nil, fmt.Errorf("not all dependency checks passed")
			}
		}
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

	var interpolatee string

	if test.Executor.Name == "manual" {
		interpolatee = test.Executor.Steps
	} else {
		interpolatee = test.Executor.Command
	}

	command, err := interpolateWithArgs(interpolatee, args, true)
	if err != nil {
		return nil, err
	}

	Println("\nExecuting test...\n")

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
		if results != "" {
			Println("****** EXECUTOR FAILED ******")
			Println(results)
			Println("*****************************")
		}

		return nil, err
	}

	Println("****** EXECUTOR RESULTS ******")
	Println(results)
	Println("******************************")

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

func GetTechnique(tid string) (*types.Atomic, error) {
	if !strings.HasPrefix(tid, "T") {
		tid = "T" + tid
	}

	var body []byte

	if LOCAL != "" {
		// Check to see if test is defined locally first. If not, body will be nil
		// and the test will be loaded below.
		body, _ = os.ReadFile(LOCAL + "/" + tid + "/" + tid + ".yaml")
		if len(body) == 0 {
			body, _ = os.ReadFile(LOCAL + "/" + tid + "/" + tid + ".yml")
		}
	}

	if len(body) == 0 {
		if BUNDLED {
			var err error

			if body, err = Technique(tid); err != nil {
				return nil, err
			}
		} else {
			orgBranch := strings.Split(REPO, "/")

			if len(orgBranch) != 2 {
				return nil, fmt.Errorf("repo must be in format <org>/<branch> (name of repo in <org> must be 'atomic-red-team')")
			}

			url := fmt.Sprintf("https://raw.githubusercontent.com/%s/atomic-red-team/%s/atomics/%s/%s.yaml", orgBranch[0], orgBranch[1], tid, tid)

			resp, err := http.Get(url)
			if err != nil {
				return nil, fmt.Errorf("getting Atomic Test from GitHub: %w", err)
			}

			defer resp.Body.Close()

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("reading Atomic Test from GitHub response: %w", err)
			}
		}
	}

	var technique types.Atomic

	if err := yaml.Unmarshal(body, &technique); err != nil {
		return nil, fmt.Errorf("processing Atomic Test YAML file: %w", err)
	}

	return &technique, nil
}

func GetMarkdown(tid string) ([]byte, error) {
	if !strings.HasPrefix(tid, "T") {
		tid = "T" + tid
	}

	var body []byte

	if LOCAL != "" {
		// Check to see if test is defined locally first. If not, body will be nil
		// and the test will be loaded below.
		body, _ = os.ReadFile(LOCAL + "/" + tid + "/" + tid + ".md")
	}

	if len(body) == 0 {
		if BUNDLED {
			var err error

			if body, err = Markdown(tid); err != nil {
				return nil, err
			}
		} else {
			orgBranch := strings.Split(REPO, "/")

			if len(orgBranch) != 2 {
				return nil, fmt.Errorf("repo must be in format <org>/<branch>")
			}

			url := fmt.Sprintf("https://raw.githubusercontent.com/%s/atomic-red-team/%s/atomics/%s/%s.md", orgBranch[0], orgBranch[1], tid, tid)

			resp, err := http.Get(url)
			if err != nil {
				return nil, fmt.Errorf("getting Atomic Test from GitHub: %w", err)
			}

			defer resp.Body.Close()

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("reading Atomic Test from GitHub response: %w", err)
			}
		}
	}

	// All the Markdown files wrap the ATT&CK technique descriptions in
	// <blockquote> blocks, but Glamour doesn't render that correctly, so let's
	// remove them here.
	body = BlockQuoteRegex.ReplaceAll(body, nil)

	return body, nil
}

func DumpTechnique(dir, tid string) (string, error) {
	if !strings.HasPrefix(tid, "T") {
		tid = "T" + tid
	}

	var (
		testBody []byte
		mdBody   []byte
	)

	// We don't check for locally defined techniques here since it makes no sense
	// to dump them to file when they're already present locally.

	if BUNDLED {
		var err error

		testBody, err = include.ReadFile("include/atomics/" + tid + "/" + tid + ".yaml")
		if err != nil {
			testBody, err = include.ReadFile("include/atomics/" + tid + "/" + tid + ".yml")
			if err != nil {
				return "", fmt.Errorf("Atomic Test is not currently bundled")
			}
		}

		mdBody, err = include.ReadFile("include/atomics/" + tid + "/" + tid + ".md")
		if err != nil {
			return "", fmt.Errorf("Atomic Test is not currently bundled")
		}
	} else {
		orgBranch := strings.Split(REPO, "/")

		if len(orgBranch) != 2 {
			return "", fmt.Errorf("repo must be in format <org>/<branch>")
		}

		url := fmt.Sprintf("https://raw.githubusercontent.com/%s/atomic-red-team/%s/atomics/%s/%s.yaml", orgBranch[0], orgBranch[1], tid, tid)

		resp, err := http.Get(url)
		if err != nil {
			return "", fmt.Errorf("getting Atomic Test from GitHub: %w", err)
		}

		defer resp.Body.Close()

		testBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("reading Atomic Test from GitHub response: %w", err)
		}

		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/atomic-red-team/%s/atomics/%s/%s.md", orgBranch[0], orgBranch[1], tid, tid)

		resp, err = http.Get(url)
		if err != nil {
			return "", fmt.Errorf("getting Atomic Test from GitHub: %w", err)
		}

		defer resp.Body.Close()

		mdBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("reading Atomic Test from GitHub response: %w", err)
		}
	}

	dir = dir + "/" + tid

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating local technique directory %s: %w", dir, err)
	}

	path := dir + "/" + tid + ".yaml"
	if err := os.WriteFile(path, testBody, 0644); err != nil {
		return "", fmt.Errorf("writing test configs for technique %s to %s: %w", tid, path, err)
	}

	path = dir + "/" + tid + ".md"
	if err := os.WriteFile(path, mdBody, 0644); err != nil {
		return "", fmt.Errorf("writing test documentation for technique %s to %s: %w", tid, path, err)
	}

	return dir, nil
}

func getTest(tid, name string, index int) (*types.AtomicTest, error) {
	Printf("\nGetting Atomic Tests technique %s from GitHub repo %s\n", tid, REPO)

	technique, err := GetTechnique(tid)
	if err != nil {
		return nil, fmt.Errorf("getting Atomic Tests technique: %w", err)
	}

	Printf("  - technique has %d tests\n", len(technique.AtomicTests))

	var test *types.AtomicTest

	if index >= 0 && index < len(technique.AtomicTests) {
		test = &technique.AtomicTests[index]
	} else {
		for _, t := range technique.AtomicTests {
			if t.Name == name {
				test = &t
				break
			}
		}
	}

	if test == nil {
		return nil, fmt.Errorf("could not find test %s/%s", tid, name)
	}

	Printf("  - found test named %s\n", test.Name)

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

	Println("\nChecking arguments...")
	Println("  - supplied on command line: " + strings.Join(keys, ", "))

	for k, v := range test.InputArugments {
		Println("  - checking for argument " + k)

		val, ok := args[k]

		if ok {
			Println("   * OK - found argument in supplied args")
		} else {
			Println("   * XX - not found, trying default arg")

			val = v.Default

			if val == "" {
				return nil, fmt.Errorf("argument [%s] is required but not set and has no default", k)
			} else {
				Println("   * OK - found argument in defaults")
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

	Printf("\nChecking platform vs our platform (%s)...\n", platform)

	var found bool

	for _, p := range test.SupportedPlatforms {
		if p == platform {
			found = true
			break
		}
	}

	if found {
		Println("  - OK - our platform is supported!")
	} else {
		return fmt.Errorf("unable to run test that supports platforms %v because we are on %s", test.SupportedPlatforms, platform)
	}

	return nil
}

func interpolateWithArgs(interpolatee string, args map[string]string, quiet bool) (string, error) {
	prevQuiet := Quiet
	Quiet = quiet

	defer func() {
		Quiet = prevQuiet
	}()

	Println("\nInterpolating command with input arguments...")

	interpolated := strings.TrimSpace(interpolatee)

	for k, v := range args {
		Printf("  - interpolating [#{%s}] => [%s]\n", k, v)

		if AtomicsFolderRegex.MatchString(v) {
			dir, err := os.MkdirTemp("", "")
			if err != nil {
				return "", fmt.Errorf("creating temp directory for %s: %w", k, err)
			}

			Println("TEMP DIR: " + dir)

			v = AtomicsFolderRegex.ReplaceAllString(v, "")
			v = strings.ReplaceAll(v, `\`, `/`)
			v = "atomics/" + v

			body, err := include.ReadFile("include/ " + v)
			if err != nil {
				return "", fmt.Errorf("reading %s: %w", k, err)
			}

			v = filepath.FromSlash(dir + "/" + v)

			if err := os.WriteFile(v, body, 0644); err != nil {
				return "", fmt.Errorf("restoring %s: %w", k, err)
			}
		}

		interpolated = strings.ReplaceAll(interpolated, "#{"+k+"}", v)
	}

	return interpolated, nil
}

func executeCommandPrompt(command string) (string, error) {
	// Printf("\nExecuting executor=cmd command=[%s]\n", command)

	cmd := exec.Command("cmd.exe", "/c", command)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing command via cmd.exe: %w", err)
	}

	return string(output), nil
}

func executeSh(command string) (string, error) {
	// Printf("\nExecuting executor=sh command=[%s]\n", command)

	cmd := exec.Command("sh", "-c", command)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing command via sh: %w", err)
	}

	return string(output), nil
}

func executeBash(command string) (string, error) {
	// Printf("\nExecuting executor=bash command=[%s]\n", command)

	cmd := exec.Command("bash", "-c", command)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing command via bash: %w", err)
	}

	return string(output), nil
}

func executePowerShell(command string) (string, error) {
	// Printf("\nExecuting executor=powershell command=[%s]\n", command)

	args := []string{"-NoProfile", command}

	cmd := exec.Command("powershell", args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing command via powershell: %w", err)
	}

	return string(output), nil
}

func executeManual(command string) (string, error) {
	// Println("\nExecuting executor=manual command=[<see below>]")

	steps := strings.Split(command, "\n")

	fmt.Printf("\nThe following steps should be executed manually:\n\n")

	for _, step := range steps {
		fmt.Printf("    %s\n", step)
	}

	return command, nil
}
