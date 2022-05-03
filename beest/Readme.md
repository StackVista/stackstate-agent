# Beest

Beest is a black-box testing framework that allow to test StackState integrations (checks and stackpacks),
by provisioning infrastructure and deploy applications in an isolated and reproducible way.

We love bees, and we drew inspiration from their world:
- bee: application
- (bee)hive: infrastructure where to place bees
- (bee)yard: set of hives with their configured bees in it
- (bee)keeper: controller from which to execute tests and intermediate actions

## Prerequisites

* Docker
* AWS account

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

By default, the state of your test yard is kept in the Terraform state which is configured to be stored in AWS S3 object.
That way Terraform ca keep track of the resources it manages, and everyone working with a given collection of 
infrastructure resources must be able to access the same state data.

For the value of the Terraform state file (therefore the S3 object key) we chose to use the name of the git branch
you are currently on (among other important variables).

In some cases if you need to spin up the same testing yard multiple times (or just want a more explicit identifier),
you can set a different value for the `RUN_ID` variable in your `.envrc`. The `RUN_ID` is not only used for the
Terraform state file name, but also used in the names of the infrastructure resources.
