## Beest

Beest is a black-box testing framework that allow to test StackState integrations (checks and stackpacks),
by provisioning infrastructure and deploy applications in an isolated and reproducible way.

We love bees, and we drew inspiration from their world:
- bee: application
- (bee)hive: infrastructure where to place bees
- (bee)yard: set of hives with their configured bees in it
- (bee)keeper: controller from which to execute tests and intermediate actions

### Prerequisites

* Docker
* AWS account

### Configure

Make a copy of envrc.example and replace the `TBD` values with your secrets:

    cp envrc.example .envrc

### Run

Start the keeper from which to run tests:

    make start
    $ beest test <scenario>
