package runner

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	StandardOutput = 1
	StandardError  = 2

	SourceSystem  = 1
	SourceCommand = 2
)

// Task describes structure of Task
type Task struct {
	// os represents operating system we are running
	os string

	// canSudo indicates whether we can use privilege elevation
	canSudo bool

	// runningAs shows user who started application
	runningAs string

	// runAs tells system which user should run this command
	runAs string

	// realtimeOutput indicates whether we want realtime output or just final output
	realtimeOutput bool

	// realtimeOutputHandler specified handler for real time output
	realtimeOutputHandler func(int, string, int)

	// failOnError will fail if any error pops up during execution (if set to true)
	failOnError bool

	// command represents command which should be executed
	command string

	// arguments represents list of arguments for command
	arguments []string

	// workingDir represents directory from which we are working
	workingDir string
}

// TaskOptions describes structure of task options used to create task
type TaskOptions struct {
	RunAs       string   `json:"run_as,omitempty"`
	Command     string   `json:"command"`
	Arguments   []string `json:"arguments,omitempty"`
	WorkingDir  string   `json:"working_dir"`
	FailOnError bool     `json:"fail_on_error,omitempty"`
}

// NewTask creates new instance of task
func NewTask(command string) *Task {
	return &Task{
		os:             runtime.GOOS,
		canSudo:        runtime.GOOS != "windows",
		runningAs:      "",
		runAs:          "",
		realtimeOutput: false,
		failOnError:    true,
		command:        command,
		arguments:      []string{},
		workingDir:     "",
	}
}

// NewTaskFromOptions creates new task from provided options
func NewTaskFromOptions(options TaskOptions) (*Task, error) {
	task := &Task{
		os:                    runtime.GOOS,
		canSudo:               runtime.GOOS != "windows",
		runningAs:             "",
		runAs:                 options.RunAs,
		realtimeOutput:        false,
		realtimeOutputHandler: nil,
		failOnError:           options.FailOnError,
		command:               options.Command,
		arguments:             options.Arguments,
		workingDir:            options.WorkingDir,
	}

	if task.runAs != "" {
		runAsTask, err := task.RunAs(task.runAs)
		if err != nil {
			return nil, err
		}
		task = runAsTask
	}

	return task, nil
}

// WithArguments sets arguments for this task
func (task *Task) WithArguments(arguments []string) *Task {
	task.arguments = arguments
	return task
}

// WithSkipError will skip error if it occurs
func (task *Task) WithSkipError() *Task {
	task.failOnError = false
	return task
}

// WithRealtimeOutput sets realtime output to a given value
func (task *Task) WithRealtimeOutput(handler func(int, string, int)) *Task {
	task.realtimeOutput = true
	task.realtimeOutputHandler = handler
	return task
}

// RunAsSudo runs with `sudo` assuming current user has access to it
func (task *Task) RunAsSudo() (*Task, error) {
	return task.RunAs("")
}

// RunAs sets user who will run this command
func (task *Task) RunAs(username string) (*Task, error) {
	if !task.canSudo {
		return nil, fmt.Errorf("this system does not support `sudo` privileges elevation")
	}

	originalUsername, err := GetMyUsername()
	if err != nil {
		return nil, err
	}
	task.runningAs = originalUsername

	if strings.TrimSpace(username) == "" {
		username = task.runningAs
	}

	canSudo, err := HasSudoPrivileges("ccm")
	if err != nil {
		return nil, err
	}

	if !canSudo {
		return nil, fmt.Errorf("user `ccm` is not in sudoers group")
	}

	task.runAs = username
	return task, nil
}

// IsSameUser checks whether we are not actually switching user
func (task *Task) IsSameUser() bool {
	return strings.TrimSpace(task.runningAs) == strings.TrimSpace(task.runAs)
}

// WithWorkingDir specifies working directory for command
func (task *Task) WithWorkingDir(directory string) *Task {
	task.workingDir = strings.TrimSpace(directory)
	return task
}

// HasWorkingDir checks whether working directory is specified
func (task *Task) HasWorkingDir() bool {
	return strings.TrimSpace(task.workingDir) != ""
}
