package driver

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/hashicorp/terraform-exec/tfexec"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultTerraformBinary = "/usr/local/bin/terraform"
)

type TerraformContext interface {
	WorkingDir() string
}

type TerraformState struct {
	Module   *tfconfig.Module
	tf       *tfexec.Terraform
	planPath string
}

func TerraformApply(ctx TerraformContext, vars map[string]string, prompt bool) {
	state := initTerraform(ctx)
	state.validate()
	if state.plan(vars) {
		state.apply(prompt)
	} else {
		log.Println("No terraform changes to apply.")
	}
}

func TerraformDestroy(ctx TerraformContext, vars map[string]string, prompt bool) {
	state := initTerraform(ctx)
	if state.plan(vars) {
		state.destroy(prompt)
	} else {
		log.Println("No terraform resource to destroy.")
	}
}

func initTerraform(ctx TerraformContext) *TerraformState {
	fmt.Printf("Initializing terraform with working dir: %s\n", ctx.WorkingDir())

	tf, err := tfexec.NewTerraform(ctx.WorkingDir(), DefaultTerraformBinary)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	// TODO debug
	//tf.SetStdout(os.Stdout)

	log.Println("Initializing terraform workspace...")
	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	return &TerraformState{
		tf:     tf,
		Module: loadModule(tf.WorkingDir()),
	}
}

func loadModule(moduleDir string) *tfconfig.Module {
	module, diags := tfconfig.LoadModule(moduleDir)
	if diags.HasErrors() {
		log.Fatalf("error reading module: %s", diags.Error())
	}
	return module
}

func (ts *TerraformState) validate() {
	log.Println("Validating terraform manifests...")
	validated, err := ts.tf.Validate(context.Background())
	if err != nil {
		log.Fatalf("error running Validate: %s", err)
	}
	if !validated.Valid {
		log.Fatalf("manifests not valid: %s", validated.Validate())
	}
}

func (ts *TerraformState) plan(vars map[string]string) bool {
	log.Println("Planning terraform changes...")
	var tfPlanOptions []tfexec.PlanOption
	for k, v := range vars {
		tfPlanOptions = append(tfPlanOptions, tfexec.Var(fmt.Sprintf("%s=%s", k, v)))
	}
	planPath := fmt.Sprintf("%s/tf.deploy", ts.Module.Path)
	tfPlanOptions = append(tfPlanOptions, tfexec.Out(planPath))

	hasChanges, err := ts.tf.Plan(context.Background(), tfPlanOptions...)
	if err != nil {
		log.Fatalf("error running Plan: %s", err)
	}
	ts.planPath = planPath
	return hasChanges
}

func (ts *TerraformState) apply(prompt bool) {
	changes, err := ts.tf.ShowPlanFileRaw(context.Background(), ts.planPath)
	if err != nil {
		log.Fatalf("error showing Plan: %s", err)
	}
	log.Println(changes)

	confirmed := true //by default, we do not ask for confirmation
	if prompt {
		log.Println("Wanna continue? [y/N]")
		confirmed = confirm()
	}

	if confirmed {
		log.Println("Applying terraform changes...")
		err = ts.tf.Apply(context.Background(), tfexec.DirOrPlan(ts.planPath))
		if err != nil {
			log.Fatalf("error running Plan: %s", err)
		} else {
			log.Println("Done applying changes.")
		}
	}
}

func (ts *TerraformState) destroy(prompt bool) {
	changes, err := ts.tf.ShowPlanFileRaw(context.Background(), ts.planPath)
	if err != nil {
		log.Fatalf("error showing Plan: %s", err)
	}
	log.Println(changes)

	confirmed := true //by default, we do not ask for confirmation
	if prompt {
		log.Println("Wanna continue? [y/N]")
		confirmed = confirm()
	}

	if confirmed {
		log.Println("Destroy terraform resources...")
		err = ts.tf.Destroy(context.Background(), tfexec.Dir(ts.planPath))
		if err != nil {
			log.Fatalf("error running Destroy: %s", err)
		} else {
			log.Println("Done destroying changes.")
		}
	}
}

func confirm() bool {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return false
	}
}

func (ts *TerraformState) state() {
	state, err := ts.tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	fmt.Println(state.FormatVersion)
}

func (ts *TerraformState) output() {
	outs, err := ts.tf.Output(context.Background())
	if err != nil {
		log.Fatalf("error running Output: %s", err)
	}
	for k, v := range outs {
		raw, err := v.Value.MarshalJSON()
		if err != nil {
			log.Fatalf("error retrieving raw value '%s': %s", k, err)
		}
		fmt.Println(fmt.Sprintf("writing output %s ...", k))
		outPath := fmt.Sprintf("%s/%s", ts.Module.Path, k)

		// output use -json option, but the out is not json, we just need to get the string and write it to a file
		strOut, err := strconv.Unquote(string(raw))
		if err != nil {
			log.Fatalf("error converting raw value: %s", outPath)
		}

		//by default we assume all outputs to be sensitive
		err = os.WriteFile(outPath, []byte(strOut), 0600)
		if err != nil {
			log.Fatalf("error writing output to: %s", outPath)
		}
	}
}

//func InitFromModule(moduleDir string) *CreationStep {
//	fmt.Printf("Initializing from module: %s\n", moduleDir)
//
//	// -from-module=MODULE-SOURCE allows to not have the .terraform directory in the same place where modules are
//	// https://www.terraform.io/docs/cli/commands/init.html#copy-a-source-module
//	//
//	// maybe we should copy modules over with rsync ?
//	//
//	err := tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.FromModule(moduleDir))
//	if err != nil {
//		log.Fatalf("error running Init: %s", err)
//	}
//
//	return &CreationStep{
//		Module: loadModule(moduleDir),
//	}
//}
//
//func init() {
//	// hardcoded TF working dir ?
//	WorkingDir := "/state"
//	fmt.Printf("Terraform working directory: %s\n", WorkingDir)
//
//	var err error
//	tf, err = tfexec.NewTerraform(WorkingDir, DefaultTerraformBinary)
//	if err != nil {
//		log.Fatalf("error running NewTerraform: %s", err)
//	}
//
//	tf.SetStdout(os.Stdout)
//}
