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
- Force Cleanup to be ran before Destroy instead of running them separately
