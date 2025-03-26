# Integration test

Integration test could be done by running the following steps.

## Install Runn

Follow the instructions in the [Runn repository](
https://github.com/k1LoW/runn
)

## Set `.env`

Set the environment variables in the `.env` file.

```env
USER_API_ENDPOINT=oqtopus-cloud.com/v1
Q_API_TOKEN=your_secret_token
DEVICE_ID=your_device_id
```

## Run all tests

```sh
task runn-all
```

## Run other tests
See the `Taskfile.yml` for the list of available tests.

```sh
task
```
