package driver

import (
	"beest/sut"
	"context"
	"fmt"
	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"log"
)

type ConnectionContext interface {
	WorkingDir() string
	Inventory() string
	PrivateKey() string
}

type AnsibleContext interface {
	ConnectionContext
	Playbook() string
	Variables() map[string]interface{}
}

func AnsiblePlay(ctx AnsibleContext) {
	vars := map[string]interface{}{
		"ansibleTasksDir": sut.AnsibleTasksPath(),
	}

	runOption := &playbook.AnsiblePlaybookOptions{
		Inventory: ctx.Inventory(),
		ExtraVars: vars,
	}
	connectionOptions := &options.AnsibleConnectionOptions{
		PrivateKey: ctx.PrivateKey(),
	}
	run := &playbook.AnsiblePlaybookCmd{
		Playbooks:         []string{ctx.Playbook()},
		Options:           runOption,
		ConnectionOptions: connectionOptions,
	}
	log.Println(fmt.Sprintf("Play Ansible playbook %s ...", run.Playbooks))
	var err = run.Run(context.Background())
	if err != nil {
		log.Fatalf("Error while preparing receiver: %s", err)
	}
	log.Println("Ansible playbook finished")
}
