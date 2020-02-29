# go-to-prod

## Summary

Tool used to validate some steps using a pipeline description.

For instance, you can validate a delivery pipeline by describing all steps that will occur on your production environment. Thanks to that tool, you will be able to automate your tests for every step before production. No more surprise during delivery! 

For now, it works only with docker & docker-compose in local. It is planned to work with:
* Remote docker-compose
* Kubernetes
* Docker-swarm
* Openshift

## How it works

This tool is a CLI application. For the moment, you have only those commands:
* check

|  Args           | Description  |
| ------------- | ----- |
| `--file=states.yml` | The file that describe a pipeline| 
| `--debug`| Enables the checker's output |

You run the command `g2p check --file=state.yml` and the tool will deploy your stack (using docker-compose) and then will do some checks using container. If the checker container exit with a code 0, then the step will be valid, else it will be failed.

If there is a failure (checker exit code != 0), then the tool will exit with code=1, otherwise it will exit with a code=0.

## How to build

`go build -o g2p`