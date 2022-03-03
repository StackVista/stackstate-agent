package driver

import (
	"beest/sut"
	"context"
	"fmt"
	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"log"
	"strings"
)

type ConnectionContext interface {
	WorkingDir() string
	Inventory() string
	PrivateKey() string
}

type AnsibleContext interface {
	ConnectionContext
	Playbook() string
	Tags() []string
	Variables() map[string]interface{}
}

func AnsiblePlay(ctx AnsibleContext) {
	vars := map[string]interface{}{
		"ansibleTasksDir": sut.AnsibleTasksPath(),
		"bees_path":       sut.BeesPath(),
	}

	runOption := &playbook.AnsiblePlaybookOptions{
		Inventory: ctx.Inventory(),
		ExtraVars: vars,
		Tags:      strings.Join(ctx.Tags(), ","),
	}
	connectionOptions := &options.AnsibleConnectionOptions{
		PrivateKey: ctx.PrivateKey(),
	}
	run := &playbook.AnsiblePlaybookCmd{
		Playbooks:         []string{ctx.Playbook()},
		Options:           runOption,
		ConnectionOptions: connectionOptions,
	}
	log.Println(fmt.Sprintf("Play Ansible playbook: %s ...", run.String()))
	var err = run.Run(context.Background())
	if err != nil {
		log.Fatalf("Error running playbook: %s", err)
	}
	log.Println("Ansible playbook finished")
}
