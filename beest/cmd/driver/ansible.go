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
	pb := fmt.Sprintf("%s/prepare.yml", step.WorkingDir())
	tags := []string{"prepare"}
	return play(step.Inventory(), pb, tags)
}

func (ad *AnsibleDeployer) Cleanup(step *step.CleanupStep) error {
	pb := fmt.Sprintf("%s/cleanup.yml", step.WorkingDir())
	tags := []string{"cleanup"}
	return play(step.Inventory(), pb, tags)
}

func play(inv, pb string, tags []string) error {
	vars := map[string]interface{}{
		"bees_path": sut.BeesPath(),
	}

	runOption := &playbook.AnsiblePlaybookOptions{
		Inventory: inv,
		ExtraVars: vars,
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
