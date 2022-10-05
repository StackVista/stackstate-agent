package cmd

import (
	"beest/cmd/driver"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(sshCmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh [scenario] [ansible-host]",
	Short: "SSH into a yard host machine",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		scenario := findScenario(args[0])
		if ansibleInventory, err := loadAnsibleInventory(scenario); err == nil {
			sshToAnsibleHost(scenario, ansibleInventory, args[1])
		}
	},
}

func loadAnsibleInventory(scenario *Scenario) (driver.AnsibleInventory, error) {
	yard := scenario.Yard.Name
	yardLocation := fmt.Sprintf("sut/yards/%s/ansible_inventory", yard)

	// Read the ansible yaml inventory file
	ansibleInventoryFile, err := ioutil.ReadFile(yardLocation)
	if err != nil {
		log.Printf("Error reading the 'ansible_inventory' for the '%s' yard at '%s'",
			yard, yardLocation)
		return driver.AnsibleInventory{}, err
	}

	// Parse the ansible yaml inventory file
	var ansibleInventory driver.AnsibleInventory
	if err := yaml.Unmarshal(ansibleInventoryFile, &ansibleInventory); err != nil {
		log.Printf("Error reading the 'ansible_inventory' for the '%s' yard at '%s'",
			yard, yardLocation)
		return driver.AnsibleInventory{}, err
	}

	return ansibleInventory, nil
}

func sshToAnsibleHost(scenario *Scenario, ansibleInventory driver.AnsibleInventory, host string) {
	yard := scenario.Yard.Name
	yardLocation := fmt.Sprintf("sut/yards/%s/ansible_inventory", yard)

	// Make sure that the host does exist in the inventory
	if inventory, ok := ansibleInventory.All.Hosts[host]; ok {
		sshCommand := fmt.Sprintf("ssh -i %s %s@%s",
			inventory.AnsibleSSHPrivateKeyFile,
			inventory.AnsibleUser,
			inventory.AnsibleHost)

		// Divert the std routing to the current std for the exec to allow user interaction
		cmd := exec.Command("bash", "-c", sshCommand)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if execErr := cmd.Run(); execErr != nil {
			log.Printf("Unable to SSH into the host '%s' for the '%s' yard at '%s'",
				host, yard, yardLocation)
			log.Printf("%s", execErr)
			return
		}

	} else {
		log.Printf("Target host '%s' does not exist in 'ansible_inventory' for the '%s' yard at '%s'",
			host, yard, yardLocation)
		return
	}
}
