package sut

import (
	"fmt"
	"log"
	"os"
)

var Cwd string

func YardPath(name string) string {
	return fmt.Sprintf("%s/sut/yards/%s", Cwd, name)
}

func BeesPath() string {
	return fmt.Sprintf("%s/sut/bees", Cwd)
}

func AnsibleTasksPath() string {
	return fmt.Sprintf("%s/sut/bees/tasks", Cwd)
}

func TestPath(groupName string) string {
	return fmt.Sprintf("%s/tests/%s", Cwd, groupName)
}

func init() {
	var err error
	Cwd, err = os.Getwd()
	if err != nil {
		log.Fatalf("Error retrieving cwd: %s", err)
	}
}
