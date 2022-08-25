package driver

import "beest/cmd/step"

// The Provisioner takes care of creating and destroying test infrastructure
type Provisioner interface {
	Create(step *step.CreationStep, prompt bool) error
	Destroy(step *step.DestroyStep, prompt bool) error
}

// The Deployer provisions applications and its configurations
type Deployer interface {
	Prepare(step *step.PrepareStep) error
	Cleanup(step *step.CleanupStep) error
}

// The Verifier runs the actual verifications
type Verifier interface {
	Verify(step *step.VerificationStep, watch bool, selection string, installDeps bool) error
}
