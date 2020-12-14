check:
	ansible windows -i hosts -m win_ping
provision:
	ansible-playbook -i hosts builder/playbook_build_env.yml
create-win-ami:
	packer build packer-agent6-winrunner.json

