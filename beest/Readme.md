# Beest

Beest is a black-box testing framework that allow to test StackState integrations (checks and stackpacks),
by provisioning infrastructure and deploy applications in an isolated and reproducible way.

We love bees, and we drew inspiration from their world:
- _bee_: application
- (bee)_hive_: infrastructure where to place bees
- (bee)_yard_: set of hives with their configured bees in it
- (bee)_keeper_: container from which to execute test steps

## Prerequisites

* Docker
* StackState
  * AWS InfoSec account
  * Artifactory account
  * License

## Configure

Make a copy of envrc.example and replace the `TBD` values with your secrets:

    cp envrc.example .envrc

## Run

Start the keeper from which to run tests:

    make

    $ beest
    ... will show the help ...

    $ beest test <scenario>
    ... will execute the full test sequence for the chosen scenario ...

Beest also support tab completion, just press `[tab]` twice after typing `beest`.

### Run state

By default, the state of your test _yard_ is kept in the Terraform state which is configured to be stored in AWS S3 object.
That way Terraform can keep track of the resources it manages, and everyone working on the same _yard_ must be able to use the same state.

For the value of the Terraform state file (therefore the S3 object key) we chose to use the name of the git branch
you are currently on (among other important variables).

In some cases if you need to spin up the same testing _yard_ multiple times (or just want a more explicit identifier),
you can set a different value for the `RUN_ID` variable in your `.envrc`. The `RUN_ID` is not only used for the
Terraform state file name, but also used in the names of the infrastructure resources.


### Feedback / Feature Requests

- Allow beast to run in the background, you can then shell in and out without the risk of closing the "make terminal window"
- Force Cleanup to be run before Destroy instead of running them separately
- Expose the scenario in the scenarios.yml file that is being ran to the wild. (Similar as with the yard_id)
  - This will allow some flexibility on selecting files in the same env and knowing the testing and root dir
- Group the .gv file + test_*-*.json file under a folder structure with the root folder the test name as seen with
  the json file and the sub folder should be date-time
- Allow the beest command to include --exclude command to ignore a certain playbook cycle
  - For example `beest prepare splunk --exclude k8s-stackstate` this will allow cli level ignoring op a playbook section
- Add functionality to reuse parts within the tests such as conftest.py instead of copying the same file


## How to setup development environment for test
We need python 3.9 for Beest.

```shell
cd beest
python3 -m venv venv
source venv/bin/activate
pip install -r ../.ci-builders/beest-base/requirements-pip-full.txt
pip install -e testframework/stscliv1/
pip install -e testframework/ststest/
```

## How to run the pytests outside of the Beest Docker Instance (IDE)
- First you need to deploy the required instances before you can run them inside your IDE. Your pytests will still use these remote resources but will execute the test from your local machine

```shell
- make
- beest create <INTEGRATION>
- beest prepare <INTEGRATION>

# Now you are ready to run things locally, you can either close Beest or leave it running in the background, the local
# execution of py files is not dependant on the Beest docker instance (I will recommend leaving the docker instance
# open so that you can kill your resource when you are done)
```

- Exit the beest docker container, or leave it open, your choice.
- Head over to the `beest/tests/<INTEGRATION>` folder and run any of the python scripts for the same `<INTEGRATION>` you ran the `create` and `prepare` for.
