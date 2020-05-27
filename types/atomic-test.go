package types

// See https://github.com/redcanaryco/atomic-red-team/blob/master/atomic_red_team/atomic_test_template.yaml

var SupportedExecutors = []string{"command_prompt", "sh", "bash", "PowerShell"}

type Atomic struct {
	AttackTechnique string       `yaml:"attack_technique"`
	DisplayName     string       `yaml:"display_name"`
	AtomicTests     []AtomicTest `yaml:"atomic_tests"`
}

type AtomicTest struct {
	Name               string   `yaml:"name"`
	Description        string   `yaml:"description"`
	SupportedPlatforms []string `yaml:"supported_platforms"`

	InputArugments map[string]InputArgument `yaml:"input_arguments"`

	DependencyExecutorName string `yaml:"dependency_executor_name,omitempty"`

	Dependencies []map[string]string `yaml:"dependencies,omitempty"`
	Executor     *AtomicExecutor     `yaml:"executor"`
}

type InputArgument struct {
	Description   string `yaml:"description"`
	Type          string `yaml:"type"`
	Default       string `yaml:"default"`
	ExpectedValue string `yaml:"expected_value"`
}

type AtomicExecutor struct {
	Name              string `yaml:"name"`
	ElevationRequired bool   `yaml:"elevation_required"`
	Command           string `yaml:"command"`
	CleanupCommand    string `yaml:"cleanup_command"`

	ExecutedCommand map[string]interface{} `yaml:"executed_command"`
}
