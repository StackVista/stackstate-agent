package driver

import (
	"context"
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
}

func AnsiblePlay(ctx AnsibleContext, vars map[string]interface{}) {
	connectionOptions := &options.AnsibleConnectionOptions{
		PrivateKey: ctx.PrivateKey(),
	}
	runOption := &playbook.AnsiblePlaybookOptions{
		Inventory: ctx.Inventory(),
		ExtraVars: vars,
	}
	run := &playbook.AnsiblePlaybookCmd{
		Playbooks:         []string{ctx.Playbook()},
		Options:           runOption,
		ConnectionOptions: connectionOptions,
	}
	var err = run.Run(context.Background())
	if err != nil {
		log.Fatalf("error while preparing receiver: %s", err)
	}
}
