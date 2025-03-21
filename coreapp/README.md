# qiqbedge
engine server

## Prerequisites
- Go (the version specified in `go.mod`)
- Task (https://taskfile.dev/)

## Commands

The commands are defined in `Taskfile.yml`.
You can choose to run the commands in the way that suits you best.

### `.env` file

The `.env` file contains environment variables used by the task runner.

```.ini
LOG_DIR=/path/to/logs
```

### `run`
Runs the engine process in the foreground.

#### Usage

```sh
task run
```

To specify a custom setting file path, use:

```sh
LOG_DIR=/path/to/log SETTING_PATH=/path/to/setting task run
```

### `build-start`
Runs the engine process in the background.

#### Usage 

```sh
task build-start
```

To specify a custom log directory and setting file path, use:

```sh
SETTING_PATH=/path/to/setting LOG_DIR=/path/to/log task build-start
```

#### Details
- First, executes `build-core` to compile the core application.
- Then, runs `{{.CORE_APP_BIN_NAME}} poller` with logs stored in `{{.LD}}` (default: `./shares/logs`) and
  settings loaded from `{{.SETTING_PATH}}` (default: `./shares/settings/setting.json`).
- metrics log path is set in settings in `{{.SETTING_PATH}}`.
- The process runs in the background via `nohup`, with output redirected to `/dev/null`.
- The process ID is stored in `{{.PID_FILE}}`

### `stop`
Stops the engine process.

#### Usage
Stops the engine process running whose process ID is stored in `{{.PID_FILE}}`.

```sh
task stop
```

## Generate codes from OAS 
### Prerequisites
- oapi-codegen
  - https://github.com/oapi-codegen/oapi-codegen
  - v2.4.1

## Generate codes
```bash
$ task generate
```

## Others
Look at the `Taskfile.yml` for more commands.