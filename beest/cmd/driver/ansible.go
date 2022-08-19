package driver

import (
	"beest/cmd/step"
	"beest/sut"
	"context"
	"fmt"
	"github.com/apenella/go-ansible/pkg/playbook"
	"log"
	"strings"
)

type AnsibleDeployer struct{}

func (ad *AnsibleDeployer) Prepare(step *step.PrepareStep) error {
	tags := []string{"prepare"}
	return play(step.Inventory(), step.Playbook(), step.Variables(), tags)
}

func (ad *AnsibleDeployer) Cleanup(step *step.CleanupStep) error {
	tags := []string{"cleanup"}
	return play(step.Inventory(), step.Playbook(), step.Variables(), tags)
}

func play(inv, pb string, vars map[string]interface{}, tags []string) error {
	ansibleVars := map[string]interface{}{
		"bees_path": sut.BeesPath(),
	}
	for k, v := range vars {
		ansibleVars[k] = v
	}

	runOption := &playbook.AnsiblePlaybookOptions{
		Inventory: inv,
		ExtraVars: ansibleVars,
		Tags:      strings.Join(tags, ","),
	}
	run := &playbook.AnsiblePlaybookCmd{
		Playbooks: []string{pb},
		Options:   runOption,
	}
	log.Println(fmt.Sprintf("Play Ansible playbook: %s ...", run.String()))
	if err := run.Run(context.Background()); err != nil {
		log.Printf("Error running playbook: %s\n", err)
		return err
	}

	log.Println("Ansible playbook finished successfully")
	return nil
}
