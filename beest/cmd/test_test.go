package cmd

import (
	"beest/cmd/driver"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type calls struct {
	created   bool
	prepared  bool
	verified  bool
	cleaned   bool
	destroyed bool
}

func TestSequence(t *testing.T) {
	// funny failures
	createFails := errors.New("create, no cigars")
	prepareFails := errors.New("prepare fails miserably")
	verifyFails := errors.New("verify won't make it")
	cleanupFails := errors.New("cleanup was an hazard")
	destroyFails := errors.New("destroy never existed")

	tests := []struct {
		name           string
		provisioner    *driver.StubProvisioner
		deployer       *driver.StubDeployer
		verifier       *driver.StubVerifier
		reset          bool
		noDestroy      bool
		expectedErrors []error
		expectedCalls  calls
	}{
		{
			name:        "all happy",
			provisioner: driver.NewStubProvisioner(nil, nil),
			deployer:    driver.NewStubDeployer(nil, nil),
			verifier:    driver.NewStubVerifier(nil),
			expectedCalls: calls{
				created:   true,
				prepared:  true,
				verified:  true,
				cleaned:   true,
				destroyed: true,
			},
		},
		{
			name:           "if create fails stop sequence",
			provisioner:    driver.NewStubProvisioner(createFails, nil),
			deployer:       driver.NewStubDeployer(nil, nil),
			verifier:       driver.NewStubVerifier(nil),
			expectedErrors: []error{createFails},
			expectedCalls: calls{
				created:   true,
				prepared:  false,
				verified:  false,
				cleaned:   false,
				destroyed: false,
			},
		},
		{
			name:           "if prepare fails do not verify, but complete sequence",
			provisioner:    driver.NewStubProvisioner(nil, nil),
			deployer:       driver.NewStubDeployer(prepareFails, nil),
			verifier:       driver.NewStubVerifier(nil),
			expectedErrors: []error{prepareFails},
			expectedCalls: calls{
				created:   true,
				prepared:  true,
				verified:  false,
				cleaned:   true,
				destroyed: true,
			},
		},
		{
			name:           "if verify fails complete sequence",
			provisioner:    driver.NewStubProvisioner(nil, nil),
			deployer:       driver.NewStubDeployer(nil, nil),
			verifier:       driver.NewStubVerifier(verifyFails),
			expectedErrors: []error{verifyFails},
			expectedCalls: calls{
				created:   true,
				prepared:  true,
				verified:  true,
				cleaned:   true,
				destroyed: true,
			},
		},
		{
			name:           "if cleanup fails complete sequence",
			provisioner:    driver.NewStubProvisioner(nil, nil),
			deployer:       driver.NewStubDeployer(nil, cleanupFails),
			verifier:       driver.NewStubVerifier(nil),
			expectedErrors: []error{cleanupFails},
			expectedCalls: calls{
				created:   true,
				prepared:  true,
				verified:  true,
				cleaned:   true,
				destroyed: true,
			},
		},
		{
			name:           "if destroy fails complete sequence",
			provisioner:    driver.NewStubProvisioner(nil, destroyFails),
			deployer:       driver.NewStubDeployer(nil, nil),
			verifier:       driver.NewStubVerifier(nil),
			expectedErrors: []error{destroyFails},
			expectedCalls: calls{
				created:   true,
				prepared:  true,
				verified:  true,
				cleaned:   true,
				destroyed: true,
			},
		},
		{
			name:           "if all steps fail except for create expect all errors back",
			provisioner:    driver.NewStubProvisioner(nil, destroyFails),
			deployer:       driver.NewStubDeployer(prepareFails, cleanupFails),
			verifier:       driver.NewStubVerifier(nil),
			expectedErrors: []error{prepareFails, cleanupFails, destroyFails},
			expectedCalls: calls{
				created:   true,
				prepared:  true,
				verified:  false,
				cleaned:   true,
				destroyed: true,
			},
		},
		{
			name:           "if reset flag and cleanup fail stop sequence",
			provisioner:    driver.NewStubProvisioner(nil, nil),
			deployer:       driver.NewStubDeployer(nil, cleanupFails),
			verifier:       driver.NewStubVerifier(nil),
			reset:          true,
			expectedErrors: []error{cleanupFails},
			expectedCalls: calls{
				created:   true,
				prepared:  false,
				verified:  false,
				cleaned:   true,
				destroyed: false,
			},
		},
		{
			name:        "if noDestroy flag do not destroy",
			provisioner: driver.NewStubProvisioner(nil, nil),
			deployer:    driver.NewStubDeployer(nil, nil),
			verifier:    driver.NewStubVerifier(nil),
			noDestroy:   true,
			expectedCalls: calls{
				created:   true,
				prepared:  true,
				verified:  true,
				cleaned:   true,
				destroyed: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario := loadScenarios().Scenarios[0]
			errs := test(tt.provisioner, tt.deployer, tt.verifier, &scenario, tt.reset, tt.noDestroy)
			assert.Equal(t, tt.expectedErrors, errs)
			assert.Equal(t, tt.expectedCalls.created, tt.provisioner.Created())
			assert.Equal(t, tt.expectedCalls.prepared, tt.deployer.Prepared())
			assert.Equal(t, tt.expectedCalls.verified, tt.verifier.Verified())
			assert.Equal(t, tt.expectedCalls.cleaned, tt.deployer.Cleaned())
			assert.Equal(t, tt.expectedCalls.destroyed, tt.provisioner.Destroyed())
		})
	}
}
