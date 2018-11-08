Agent Molecule tests
--------------------

Those are integration tests that spawn new VMs in AWS and do the following:

* install the agent from the debian repository
* run a docker compose setup of the StackState receiver
* verify assertion on the target VMs

### Run

From the parent directory execute `./run-melecule.sh `

Molecule is based on the following lifecycle:

