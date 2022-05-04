# Runner Package for GoLang
This package provides easy way to run commands on host with support for realtime output.

# Limitations
Since this package leverages `sudo` functionality only available on Unix systems, it will not work for binaries built for Windows.  
Well, it will, but you will not be able to execute any commands which require privileges escalation.

# Runner Tasks
Tasks are jobs which runner must perform. Think of them as commands on steroids.  
Tasks are passed to the `Run` method and executed. There are many options you can tweak.  
There are two ways to create new task, one is by using task builder, the other is by using task options.

## Creating new task with task builder
Task builder allows you to have granular control of what is going to be executed, as you are the one who is building the logic for the task.
This method should be used only when you feel like you need more control over the settings and your task configuration is mostly "static".
```go
task, err := runner.
    NewTask("apt").
    WithArguments([]string{"update"}).
    WithRealtimeOutput(func (outputType int, outputLine string, outputSource int) {
        fmt.Println(fmt.Sprintf("[type: %d, source: %d] %s", outputType, outputSource, outputLine))
    }).
    RunAsSudo()
```

1. `NewTask` - This will create new task, which will have `apt` as a command
2. `WithArguments` - Then we are passing `update` as a list of arguments
   1. Right now, complete command looks like `apt update`
3. `WithRealtimeOutput` - allows you to return output in real time to the specified handler
4. `RunAsSudo` - will execute command as `sudo`
   1. Worth noting, `RunAs` and `RunAsSudo` as completely different commands.
   2. `RunAsSudo` will execute command with `sudo` prefix ONLY if user who started application has access to SUDO.
   3. Right now, complete command looks like `sudo apt update`

## Creating new task with task options
Task options should be used whenever you have multiple different commands which can be configured remotely.
```go
task, err := runner.NewTaskFromOptions(runner.TaskOptions{
	RunAs:     "cabinet",
	Command:   "apt",
	Arguments: []string{"update"},
})
if err != nil {
	fmt.Println(err)
	return
}

task.WithRealtimeOutput(func (outputType int, outputLine string, outputSource int) {
    fmt.Println(fmt.Sprintf("[type: %d, source: %d] %s", outputType, outputSource, outputLine))
})
```

The procedure is pretty much the same as for when you are creating task with task builder.  
Only difference is, `NewTaskFromOptions` returns `error` as second parameter.  
This is due to the fact that we are handling privilege escalation inside this method, so you don't have to call `RunAs` afterwards.  
You still have to call `RunAsSudo` if you want to `sudo` from current user.

# Available Task methods

## NewTask
Creates new instance of task
```go
func NewTask(command string) *Task {}
```

## NewTaskFromOptions
Creates new task from provided options
```go
func NewTaskFromOptions(options TaskOptions) (*Task, error) {}
```

## WithArguments
Sets arguments for this task
```go
func (task *Task) WithArguments(arguments []string) *Task {}
```

## WithSkipError
Will skip error if it occurs
```go
func (task *Task) WithSkipError() *Task {}
```

## WithRealtimeOutput
Enable realtime output and specify handler
```go
func (task *Task) WithRealtimeOutput(handler func(string, string)) *Task {}
```

## RunAsSudo
Runs with `sudo` assuming current user has access to it
```go
func (task *Task) RunAsSudo() (*Task, error) {}
```

## RunAs
Sets user who will run this command
```go
func (task *Task) RunAs(username string) (*Task, error) {}
```

# Running created task
To run created task, you need to call ```Run()``` method.
```go
executionResult, err := runner.Run(task)
```

`executionResult` - will be `true` or `false`, depending on execution status
`err` - will be instance of `error` or `nil`, depending on execution status

## Running async
To run task in a separate thread, simply run this task in goroutine.
```go
go runner.Run(task)
```