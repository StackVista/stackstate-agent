package step

import "fmt"

func (h *Hive) Inventory() string {
	return fmt.Sprintf("%s/ansible_inventory", h.dir)
}

func (h *Hive) PrivateKey() string {
	return fmt.Sprintf("%s/id_rsa", h.dir)
}

///

type PrepareStep struct {
	Hive
}

func Prepare(create *CreationStep) *PrepareStep {
	return &PrepareStep{Hive{create.dir}}
}

func (p *PrepareStep) Playbook() string {
	return fmt.Sprintf("%s/prepare.yml", p.dir)
}

///

type VerificationStep struct {
	Hive
	hosts []string
}

func Verify(prepare *PrepareStep, hosts []string) *VerificationStep {
	return &VerificationStep{
		Hive{prepare.dir},
		hosts,
	}
}

func (v *VerificationStep) Hostnames() []string {
	return v.hosts
}

///

type CleanupStep struct {
	Hive
}

func Cleanup(prepare *PrepareStep) *CleanupStep {
	return &CleanupStep{Hive{prepare.dir}}
}

func (c *CleanupStep) Playbook() string {
	return fmt.Sprintf("%s/cleanup.yml", c.dir)
}
