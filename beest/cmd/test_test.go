package cmd

import (
	"beest/cmd/driver"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

//func TestRunCleanupWhenPrepareFails(t *testing.T) {
//	// create stub provisioner which succeed for create and destroy
//	stubProvisioner := driver.NewStubProvisioner(nil, nil)
//	// create stub deployer which fails prepare but succeed cleanup
//	expectedPrepareFailure := errors.New("prepare fails miserably")
//	stubDeployer := driver.NewStubDeployer(expectedPrepareFailure, nil)
//	// create stub verifier which will not be called
//	stubVerifier := driver.NewStubVerifier(nil)
//	// given a scenario
//	scenario := loadScenarios().Scenarios[0]
//
//	// run full test sequence
//	errs := test(stubProvisioner, stubDeployer, stubVerifier, &scenario, false, false)
//
//	assert.Len(t, errs, 1)
//	assert.ErrorIs(t, errs[0], expectedPrepareFailure)
//
//	assert.Equal(t, true, stubProvisioner.Created(), "not created")
//	assert.Equal(t, true, stubDeployer.Prepared(), "not prepared")
//	assert.Equal(t, false, stubVerifier.Verified(), "verified")
//	assert.Equal(t, true, stubDeployer.Cleaned(), "not cleaned")
//	assert.Equal(t, true, stubProvisioner.Destroyed(), "not destroyed")
//}
//
//func TestPropagateVerifyError(t *testing.T) {
//	stubProvisioner := driver.NewStubProvisioner(nil, nil)
//	stubDeployer := driver.NewStubDeployer(nil, nil)
//	expectedVerifyFailure := errors.New("verify fails miserably")
//	stubVerifier := driver.NewStubVerifier(expectedVerifyFailure)
//
//	scenario := loadScenarios().Scenarios[0]
//
//	// run full test sequence
//	errs := test(stubProvisioner, stubDeployer, stubVerifier, &scenario, false, false)
//
//	assert.Len(t, errs, 1)
//	assert.ErrorIs(t, errs[1], expectedVerifyFailure)
//
//	assert.Equal(t, true, stubProvisioner.Created(), "not created")
//	assert.Equal(t, true, stubDeployer.Prepared(), "not prepared")
//	assert.Equal(t, true, stubVerifier.Verified(), "not verified")
//	assert.Equal(t, true, stubDeployer.Cleaned(), "not cleaned")
//	assert.Equal(t, true, stubProvisioner.Destroyed(), "not destroyed")
//}

type calls struct {
	created   bool
	prepared  bool
	verified  bool
	cleaned   bool
	destroyed bool
}

func TestSequence(t *testing.T) {
	tests := []struct {
		name        string
		provisioner *driver.StubProvisioner
		deployer    *driver.StubDeployer
		verifier    *driver.StubVerifier
		errors      []error
		calls       calls
	}{
		{
			name:        "cleanup should run when prepare fails",
			provisioner: driver.NewStubProvisioner(nil, nil),
			deployer:    driver.NewStubDeployer(errors.New("prepare fails miserably"), nil),
			verifier:    driver.NewStubVerifier(nil),
			errors:      []error{errors.New("prepare fails miserably")},
			calls: calls{
				created:   true,
				prepared:  true,
				verified:  false,
				cleaned:   true,
				destroyed: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario := loadScenarios().Scenarios[0]
			errs := test(tt.provisioner, tt.deployer, tt.verifier, &scenario, false, false)
			assert.Equal(t, tt.errors, errs)
			assert.Equal(t, tt.calls.created, tt.provisioner.Created())
			assert.Equal(t, tt.calls.prepared, tt.deployer.Prepared())
			assert.Equal(t, tt.calls.verified, tt.verifier.Verified())
			assert.Equal(t, tt.calls.cleaned, tt.deployer.Cleaned())
			assert.Equal(t, tt.calls.destroyed, tt.provisioner.Destroyed())
		})
	}
}
