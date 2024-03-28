# RUN THESE MAKE COMMANDS FROM THE CORE BEEST DIRECTORY.
# IT IS INCLUDE IN THE PRIMARY MAKEFILE BUT SPLIT TO KEEP EVERYTHING RELATED ON ONE PLACE
# SHORTCUT COMMANDS FOR A FASTER EXECUTION CYCLE FOR SPLUNK

SHELL_INIT := source .envrc; ./bootstrap.sh

create_splunk:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest create splunk'

prepare_splunk:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest prepare splunk'

verify_splunk:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest verify splunk'

remove_splunk:
	docker exec -ti ${KEEPER_CONTAINER} bash -c '$(SHELL_INIT); ./beest cleanup splunk; ./beest destroy splunk'
