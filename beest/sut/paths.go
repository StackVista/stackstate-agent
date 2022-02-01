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
		log.Fatalf("Error while retrieving cwd: %s", err)
	}
}
