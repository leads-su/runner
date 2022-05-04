package runner

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/leads-su/logger"
)

// Runner describes structure of runner
type Runner struct {
	task      *Task
	onError   func()
	onSuccess func()
}

// RunnerOptions describes options which can be passed as runner configuration
type RunnerOptions struct {
	Task      *Task
	OnError   func()
	OnSuccess func()
}

// NewRunner creates new instance of runner
func NewRunner(options *RunnerOptions) *Runner {
	return &Runner{
		task:      options.Task,
		onError:   options.OnError,
		onSuccess: options.OnSuccess,
	}
}

// Run old method which was used before runner creation mechanism
func Run(task *Task) (bool, error) {
	return NewRunner(&RunnerOptions{
		Task: task,
	}).Run()
}

// Run executes task given in runner configuration
func (runner *Runner) Run() (bool, error) {
	var executionArray []string

	navigationCommand := ""
	taskCommand := fmt.Sprintf("%s %s", runner.task.command, strings.Join(runner.task.arguments, " "))

	if runner.task.HasWorkingDir() {
		navigationCommand = fmt.Sprintf("cd %s", runner.task.workingDir)
	}

	if navigationCommand != "" {
		executionArray = append(executionArray, navigationCommand)
	}
	executionArray = append(executionArray, taskCommand)

	cmd, arguments := runner.baseCommandBuilder()
	arguments = append(arguments, strings.Join(executionArray, " && "))

	command := exec.Command(cmd, arguments...)

	if !runner.task.realtimeOutput {
		_, stdErr := command.Output()
		if stdErr != nil {
			if runner.onError != nil {
				go runner.onError()
			}
			logger.Errorf("runner:runner", fmt.Sprintf("failed to execute command - %s", stdErr.Error()))
			return false, stdErr
		}

		if runner.onSuccess != nil {
			go runner.onSuccess()
		}
		return true, nil
	}

	stdOut, err := command.StdoutPipe()
	if err != nil {
		return runner.handleRealtimeErrorCase("failed to create stdout pipe - %s", err)
	}

	stdErr, err := command.StderrPipe()
	if err != nil {
		return runner.handleRealtimeErrorCase("failed to create stderr pipe - %s", err)
	}

	err = command.Start()
	if err != nil {
		return runner.handleRealtimeErrorCase("failed to start command - %s", err)
	}

	stdOutBuffer := bufio.NewScanner(stdOut)
	stdErrBuffer := bufio.NewScanner(stdErr)

	for stdOutBuffer.Scan() {
		result, err := runner.processScannedLine(stdOutBuffer, StandardOutput)
		if !result {
			result, err = runner.handleRealtimeErrorCase("failed to process `stdout` line - %s", err)
		}
		if err != nil && runner.task.failOnError {
			return result, err
		}
	}

	for stdErrBuffer.Scan() {
		result, err := runner.processScannedLine(stdErrBuffer, StandardError)
		if !result {
			result, err = runner.handleRealtimeErrorCase("failed to process `stderr` line - %s", err)
		}
		if err != nil && runner.task.failOnError {
			return result, err
		}
	}

	err = command.Wait()
	if err != nil {
		return runner.handleRealtimeErrorCase("command failed with following error - %s", err)
	}
	return runner.handleRealtimeSuccessCase(command)
}

// baseCommandBuilder builds command which will be used for execution
func (runner *Runner) baseCommandBuilder() (string, []string) {
	command := "/bin/bash"
	arguments := []string{"-lc"}

	if runner.task.runAs != "" {
		if !runner.task.IsSameUser() {
			arguments = append([]string{"-HSu", runner.task.runAs, command}, arguments...)
			command = "sudo"
		}
	}

	return command, arguments
}

// processScannedLine handles scanned line
func (runner *Runner) processScannedLine(scanner *bufio.Scanner, outputType int) (bool, error) {
	err := scanner.Err()
	if err != nil {
		return false, err
	}

	line := scanner.Text()
	if outputType == StandardError && line != "" {
		runner.task.realtimeOutputHandler(outputType, line, SourceCommand)
	} else if outputType == StandardOutput {
		runner.task.realtimeOutputHandler(outputType, line, SourceCommand)
	}
	return true, nil
}

// handleRealtimeErrorCase handles ERROR case for Realtime output
func (runner *Runner) handleRealtimeErrorCase(format string, err error) (bool, error) {
	logger.Errorf("runner:runner", fmt.Sprintf(format, err.Error()))
	if runner.onError != nil {
		go runner.onError()
	}
	runner.task.realtimeOutputHandler(StandardError, err.Error(), SourceSystem)
	return false, err
}

// handleRealtimeSuccessCase handles SUCCESS case for Realtime output
func (runner *Runner) handleRealtimeSuccessCase(command *exec.Cmd) (bool, error) {
	successMessage := fmt.Sprintf("exit status %d", command.ProcessState.ExitCode())
	runner.task.realtimeOutputHandler(StandardOutput, successMessage, SourceSystem)
	if runner.onSuccess != nil {
		go runner.onSuccess()
	}
	return true, nil
}
