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

func (ad *AnsibleDeployer) Prepare(step *step.PrepareStep, exclusions []string) error {
	tags := []string{"prepare"}
	return play(step.Inventory(), step.Playbook(), step.Yard.TestingDir(), tags, exclusions)
}

func (ad *AnsibleDeployer) Cleanup(step *step.CleanupStep, exclusions []string) error {
	tags := []string{"cleanup"}
	return play(step.Inventory(), step.Playbook(), step.Yard.TestingDir(), tags, exclusions)
}

func play(inv, pb, testingDir string, tags []string, exclusions []string) error {
	vars := map[string]interface{}{
		"bees_path":           sut.BeesPath(),
		"scenario_test_group": testingDir,
	}

	runOption := &playbook.AnsiblePlaybookOptions{
		Inventory: inv,
		ExtraVars: vars,
		Tags:      strings.Join(tags, ","),
		SkipTags:  strings.Join(exclusions, ","),
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
