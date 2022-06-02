package types

// See https://github.com/redcanaryco/atomic-red-team/blob/master/atomic_red_team/atomic_test_template.yaml
// and https://github.com/redcanaryco/atomic-red-team/blob/master/atomic_red_team/spec.yaml

var SupportedExecutors = []string{"bash", "command_prompt", "manual", "powershell", "sh"}

type Atomic struct {
	AttackTechnique string       `yaml:"attack_technique"`
	DisplayName     string       `yaml:"display_name"`
	AtomicTests     []AtomicTest `yaml:"atomic_tests"`

	BaseDir string `yaml:"-"`
}

type AtomicTest struct {
	Name               string   `yaml:"name"`
	GUID               string   `yaml:"auto_generated_guid,omitempty"`
	Description        string   `yaml:"description,omitempty"`
	SupportedPlatforms []string `yaml:"supported_platforms"`

	InputArugments map[string]InputArgument `yaml:"input_arguments,omitempty"`

	DependencyExecutorName string `yaml:"dependency_executor_name,omitempty"`

	Dependencies []Dependency    `yaml:"dependencies,omitempty"`
	Executor     *AtomicExecutor `yaml:"executor"`

	BaseDir string `yaml:"-"`
}

type InputArgument struct {
	Description   string `yaml:"description"`
	Type          string `yaml:"type"`
	Default       string `yaml:"default"`
	ExpectedValue string `yaml:"expected_value,omitempty"`
}

type Dependency struct {
	Description      string `yaml:"description"`
	PrereqCommand    string `yaml:"prereq_command,omitempty"`
	GetPrereqCommand string `yaml:"get_prereq_command,omitempty"`
}

type AtomicExecutor struct {
	Name              string `yaml:"name"`
	ElevationRequired bool   `yaml:"elevation_required"`
	Command           string `yaml:"command,omitempty"`
	Steps             string `yaml:"steps,omitempty"`
	CleanupCommand    string `yaml:"cleanup_command,omitempty"`

	ExecutedCommand map[string]interface{} `yaml:"executed_command,omitempty"`
}
