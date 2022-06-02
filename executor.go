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

func Execute(tid, name string, index int, inputs []string, env []string) (*types.AtomicTest, error) {
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

	if env == nil {
		Println(" Env:       <none>")
	} else {
		Println(" Env:       " + strings.Join(env, "\n            "))
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

			command, err := interpolateWithArgs(dep.PrereqCommand, test.BaseDir, args, true)
			if err != nil {
				return nil, err
			}

			switch test.DependencyExecutorName {
			case "bash":
				_, err = executeBash(command, env)
			case "command_prompt":
				_, err = executeCommandPrompt(command, env)
			case "manual":
				_, err = executeManual(command)
			case "powershell":
				_, err = executePowerShell(command, env)
			case "sh":
				_, err = executeSh(command, env)
			}

			if err == nil {
				Printf("   * OK - dependency check succeeded!\n")
				continue
			}

			command, err = interpolateWithArgs(dep.GetPrereqCommand, test.BaseDir, args, true)
			if err != nil {
				return nil, err
			}

			var result string

			switch test.DependencyExecutorName {
			case "bash":
				result, err = executeBash(command, env)
			case "command_prompt":
				result, err = executeCommandPrompt(command, env)
			case "manual":
				result, err = executeManual(command)
			case "powershell":
				result, err = executePowerShell(command, env)
			case "sh":
				result, err = executeSh(command, env)
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

	command, err := interpolateWithArgs(interpolatee, test.BaseDir, args, false)
	if err != nil {
		return nil, err
	}

	Println("\nExecuting test...\n")

	var results string

	switch test.Executor.Name {
	case "bash":
		results, err = executeBash(command, env)
	case "command_prompt":
		results, err = executeCommandPrompt(command, env)
	case "manual":
		results, err = executeManual(command)
	case "powershell":
		results, err = executePowerShell(command, env)
	case "sh":
		results, err = executeSh(command, env)
	}

	if err != nil {
		if results != "" {
			Println("****** EXECUTOR FAILED ******")
			Println(results)
			Println("*****************************")
		}

		return nil, err
	}

	if test.Executor.Name != "manual" {
		Println("****** EXECUTOR RESULTS ******")
		Println(results)
		Println("******************************")
	}

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

	if len(body) != 0 {
		var technique types.Atomic

		if err := yaml.Unmarshal(body, &technique); err != nil {
			return nil, fmt.Errorf("processing Atomic Test YAML file: %w", err)
		}

		technique.BaseDir = LOCAL
		return &technique, nil
	}

	if BUNDLED {
		body, base, err := Technique(tid)
		if err != nil {
			return nil, err
		}

		var technique types.Atomic

		if err := yaml.Unmarshal(body, &technique); err != nil {
			return nil, fmt.Errorf("processing Atomic Test YAML file: %w", err)
		}

		technique.BaseDir = base
		return &technique, nil
	}

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

	// TODO: also dump any additional files bundled with techniques.

	if BUNDLED {
		var err error

		if testBody, _, err = Technique(tid); err != nil {
			return "", err
		}

		if mdBody, err = Markdown(tid); err != nil {
			return "", err
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

	test.BaseDir = technique.BaseDir

	Printf("  - found test named %s\n", test.Name)

	return test, nil
}

func checkArgsAndGetDefaults(test *types.AtomicTest, inputs []string) (map[string]string, error) {
	var (
		keys    []string
		args    = make(map[string]string)
		updated = make(map[string]string)
	)

	if len(test.InputArugments) == 0 {
		return updated, nil
	}

	for _, i := range inputs {
		kv := strings.Split(i, "=")

		if len(kv) == 2 {
			keys = append(keys, kv[0])
			args[kv[0]] = kv[1]
		}
	}

	Println("\nChecking arguments...")

	if len(keys) > 0 {
		Println("  - supplied on command line: " + strings.Join(keys, ", "))
	}

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

func interpolateWithArgs(interpolatee, base string, args map[string]string, quiet bool) (string, error) {
	interpolated := strings.TrimSpace(interpolatee)

	if len(args) == 0 {
		return interpolated, nil
	}

	prevQuiet := Quiet
	Quiet = quiet

	defer func() {
		Quiet = prevQuiet
	}()

	Println("\nInterpolating command with input arguments...")

	for k, v := range args {
		Printf("  - interpolating [#{%s}] => [%s]\n", k, v)

		if AtomicsFolderRegex.MatchString(v) {
			v = AtomicsFolderRegex.ReplaceAllString(v, "")
			v = strings.ReplaceAll(v, `\`, `/`)
			v = strings.TrimSuffix(base, "/") + "/" + v

			// TODO: handle requesting file from GitHub repo if not bundled.
			if base != LOCAL {
				body, err := include.ReadFile(v)
				if err != nil {
					return "", fmt.Errorf("reading %s: %w", k, err)
				}

				v = filepath.FromSlash(TEMPDIR + "/" + v)

				if err := os.MkdirAll(filepath.Dir(v), 0700); err != nil {
					return "", fmt.Errorf("creating directory structure for %s: %w", k, err)
				}

				if err := os.WriteFile(v, body, 0644); err != nil {
					return "", fmt.Errorf("restoring %s: %w", k, err)
				}
			}
		}

		interpolated = strings.ReplaceAll(interpolated, "#{"+k+"}", v)
	}

	return interpolated, nil
}

func executeCommandPrompt(command string, env []string) (string, error) {
	// Printf("\nExecuting executor=cmd command=[%s]\n", command)

	f, err := os.Create(TEMPDIR + "/goart.bat")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}

	if _, err := f.Write([]byte(command)); err != nil {
		f.Close()

		return "", fmt.Errorf("writing command to file: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing batch file: %w", err)
	}

	cmd := exec.Command("cmd.exe", "/c", f.Name())
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing batch file: %w", err)
	}

	return string(output), nil
}

func executeSh(command string, env []string) (string, error) {
	// Printf("\nExecuting executor=sh command=[%s]\n", command)

	f, err := os.Create(TEMPDIR + "/goart.sh")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}

	if _, err := f.Write([]byte(command)); err != nil {
		f.Close()

		return "", fmt.Errorf("writing command to file: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing shell script: %w", err)
	}

	cmd := exec.Command("sh", f.Name())
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing shell script: %w", err)
	}

	return string(output), nil
}

func executeBash(command string, env []string) (string, error) {
	// Printf("\nExecuting executor=bash command=[%s]\n", command)

	f, err := os.Create(TEMPDIR + "/goart.bash")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}

	if _, err := f.Write([]byte(command)); err != nil {
		f.Close()

		return "", fmt.Errorf("writing command to file: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing bash script: %w", err)
	}

	cmd := exec.Command("bash", f.Name())
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing bash script: %w", err)
	}

	return string(output), nil
}

func executePowerShell(command string, env []string) (string, error) {
	// Printf("\nExecuting executor=powershell command=[%s]\n", command)

	f, err := os.Create(TEMPDIR + "/goart.ps1")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}

	if _, err := f.Write([]byte(command)); err != nil {
		f.Close()

		return "", fmt.Errorf("writing command to file: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing PowerShell script: %w", err)
	}

	cmd := exec.Command("powershell", "-NoProfile", f.Name())
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("executing PowerShell script: %w", err)
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
