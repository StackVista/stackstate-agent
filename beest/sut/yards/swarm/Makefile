# RUN THESE MAKE COMMANDS FROM THE CORE BEEST DIRECTORY.
# IT IS INCLUDE IN THE PRIMARY MAKEFILE BUT SPLIT TO KEEP EVERYTHING RELATED ON ONE PLACE
# SHORTCUT COMMANDS FOR A FASTER EXECUTION CYCLE FOR SWARM

SHELL_INIT := source .envrc; ./bootstrap.sh

create_swarm:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest create swarm'

prepare_swarm:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest prepare swarm'

verify_swarm:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest verify swarm'

remove_swarm:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest cleanup swarm; ./beest destroy swarm'
