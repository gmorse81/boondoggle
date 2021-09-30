[![Build Status](https://cloud.drone.io/api/badges/gmorse81/boondoggle/status.svg)](https://cloud.drone.io/gmorse81/boondoggle) 
# Boondoggle

Boondoggle is a helm umbrella chart preprocessor and a dependency state management tool as well as a local development tool.

## What is an umbrella chart?

If you are unfamiliar with the concept of an umbrella chart, please see [the helm recommendation for managing complex charts](https://helm.sh/docs/howto/charts_tips_and_tricks/#complex-charts-with-many-dependencies)

tldr; An umbrella chart is a helm chart that ties together a number of other charts as dependencies.

## What problem is being solved?

If all of the subcharts in your umbrella are from public sources and are not written by you, you may not need boondoggle. However, if you have created subcharts that are independent projects under active development, you will likely find boondoggle useful.

The problem is that dependencies defined by an umbrella chart are generally pulled from a chart repository when being run in production or QA, but for local development, they need to be run locally. You can edit the requirements of the umbrella to do this, but then you take the risk of having that code committed to your project and having something fail.

## How does it solve the problem?

Boondoggle solves this by defining the different environments and with their dependencies and states as defined in the umbrella chart.  it does this by dynamically generating the umbrella project requirements.yaml. Beyond managing the requirments.yaml file of the helm chart, it also supports the ability to run a dependency locally by cloning it from github, building the container image and specifying files and values to the helm command that installs the deployment locally.  It solves the boondoggle of state management for local dev (hence the name).

## How does it work?

Boondoggle defines a boondoggle.yml file that contains the configuration for the different environments and states of its subcharts. It also accepts flags to specify the state and environment.

Here's an annotated boondoggle.yml file:

```yaml
# when specified, boondoggle will promt you for your docker hub credentials then create 
# a kubernetes docker-registry type secret.
#
# eg kubectl create secret docker-registry dockerregcreds
pull-secrets-name: dockerregcreds
# Specify 2 or 3 (default is 2) so boondoggle knows which version of helm command syntax to use.
helmVersion: 3
# when specified, boondoggle will add the following helm chart repos. if promptbasicauth is true, 
# it will ask for a username and password.
helm-repos:
  - name: stable
    url: https://kubernetes-charts.storage.googleapis.com
  - name: myprivaterepo
    url: https://privaterepo.example.com/repo
    promptbasicauth: true
    # You can provide the username and password in boondoggle.yml. Both support 
    # environment variable replacement. 
    # If the replaced value is an empty string, it will fall back to prompting for the values.
    # This is useful if you are running boondoggle in an automated fashion.
    username: myrepousername
    password: ${HELM_PASS}
# Specify the details and path of your umbrella chart relative to the boondoggle.yml file.
umbrella:
  # name of the chart as specified in chart.yaml
  name: mwg-umbrella-chart
  # relative path to the chart
  path: example-umbrella
  # specify different environments in which the umbrella could be configured
  environments:
    # the dev environment
    - name: dev
      # files are appended to the helm command like this: -f ${pwd}/PATH/local.yml
      files:
        - "local.yml"
      # values are appended to the helm command like this: --set global.localenv=true
      # use of environment vars is supported. eg. - "global.thisdir=${PWD}"
      values:
        - "global.localenv=true"
    - name: test
      files:
        - "test.yml"
    # the default environment. This environment will be used if no flags are provided to 
    # the boondoggle command.
    - name: default
      files:
        - "prod.yml"
# Specify the services that will be used in the requirements.yaml file in the umbrella chart.
services:
    # Specify the name of the dependency (best practice is to use the name of the git repo in 
    # which this dependency lives)
  - name: my-dependency
    # Specify the path to this dependency relative to the boondoggle.yml file. This will be 
    # used to clone the project if specified.
    path: source-projects/my-dependency
    # git repo of chart for reference
    gitrepo: git@github.com:my-account/my-dependency.git
    # The alias of the dependency as to be used in requirements.yaml
    alias: awesome-chart
    # The name of the chart as specified in this project's chart.yml. note: boondoggle expects 
    # this chart to live at PATH/CHART
    chart: my-dependency-chart
    # Specify any number of states
    states:
      # This state is called "local"
      - state-name: local
        # If specified, and the repository name is "localdev", these commands will be run before the helm deployment.
        # use of environment vars is supported. 
        # eg. - "build -t myaccount/myimage:${MY_CI_TAG} source-projects/my-dependency/."
        preDeploySteps:
          - cmd: docker
            args: ["build", "-t", "myaccount/myimage:dev", "source-projects/my-dependency/."]
        # If specified, and the repository name is "localdev", these commands will be run after the helm deployment.
        postDeploySteps:
          - cmd: docker
            args: ["run", "--rm", "-t", "-v", "${PWD}/source-projects/my-dependency:/usr/src/app", "-e", "NODE_ENV=development", "myaccount/myimage:dev", "npm", "install"]
        # If specified, and the repository name is "localdev", these commands will be executed inside the container. 
        # It will find a pod with an "app" label matching the app value, then exec the command specified in "args" 
        # against the container specified in "container".
        postDeployExec:
          - app: my-pod-app-name
            container: my-container-name
            args: ["touch", "/testfile"]
        # Values passe to the helm install command like this: --set awesome-chart.localdev=true 
        # note that the alias or chart value is prepended to the value automatically by boondoggle
        # use of environment vars is supported. eg. - "thisdir=${PWD}"
        helm-values:
          - "localdev=true"
        # The version of the chart as specified in requirements.yaml
        version: x
        # The helm chart repo. Very important note: if you specify "localdev" as the repo, boondoggle will use your 
        # local code like this file:///${PWD}/PATH/CHART
        repository: localdev
        # the default state. using the word "default" here indicated to boondoggle that if no other flags are supplied 
        # to the command, this is the state you want to use.
      - state-name: default
        repository: "@myprivaterepo"
        version: ~1
  # Another service. Same as above.
  - name: dev-mysql
    path: source-projects/dev-mysql
    gitrepo: git@github.com:myusername/dev-mysql.git
    alias: mwg-dev-mysql
    chart: dev-mysql-chart
    dep-values-all-states:
      tags:
        - "dev"
    states:
      - state-name: local
        helm-values:
          - "localdev=true"
        version: x
        repository: localdev
      - state-name: default
        repository: "@myprivaterepo"
        version: ~1
```

## How do you use it?

`boondoggle --version` will now output the version number.

After creating your boondoggle.yml and placing it into the git repo which contains your helm umbrella, run the `boondoggle up` command with any flags you need to specify environemt and state.

here's the output from boondoggle up --help

    boondoggle up with no extra flags will configure your defaults and deploy using helm.
    Flags can be used to change configuration based on your needs.

    Usage:
    boondoggle up [flags]

    Flags:
    -h, --help               help for up
        --namespace string   The kubernetes namespace of this release
        --release string     The helm release name

    Global Flags:
        --config string                  config file (./boondoggle.yml)
    -d, --dry-run                        Dry run will do all steps except for the helm deploy. The helm command that would have been run will be printed.
    -v, --verbose                        Verbose output.
        --supersecret                    Will output EVERYTHING. Charts, secrets, keys, certificates. Everything.
    -e, --environment string             Selects the umbrella environment. Defaults to the environment with name: default in the boondoggle.yml file. (default "default")
    -s, --service-state stringSlice      Sets a services state eg. my-service=local. Defaults to the 'default' state.
    -a, --set-state-all string           Sets all services to the same state.
    -k, --skip-docker                    Skips the docker build step.
    -o, --state-v-override stringSlice   Override a services's version for the state specified. eg. my-service=1.0.0
