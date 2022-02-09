package step

import "fmt"

func (yard *Yard) Inventory() string {
	return fmt.Sprintf("%s/ansible_inventory", yard.dir)
}

func (yard *Yard) PrivateKey() string {
	return fmt.Sprintf("%s/id_rsa", yard.dir)
}

///

type PrepareStep struct {
	Yard
}

func Prepare(create *CreationStep) *PrepareStep {
	return &PrepareStep{create.Yard}
}

func (p *PrepareStep) Playbook() string {
	return fmt.Sprintf("%s/prepare.yml", p.dir)
}

///

type VerificationStep struct {
	Yard
	testsPath string
	hosts     []string
}

func Verify(prepare *PrepareStep, testsPath string, hosts []string) *VerificationStep {
	return &VerificationStep{
		prepare.Yard,
		testsPath,
		hosts,
	}
}

func (v *VerificationStep) TestsPath() string {
	return v.testsPath
}

func (v *VerificationStep) Hostnames() []string {
	return v.hosts
}

///

type CleanupStep struct {
	Yard
}

func Cleanup(prepare *PrepareStep) *CleanupStep {
	return &CleanupStep{prepare.Yard}
}

func (c *CleanupStep) Playbook() string {
	return fmt.Sprintf("%s/cleanup.yml", c.dir)
}
