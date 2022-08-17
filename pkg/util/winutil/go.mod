module github.com/StackVista/stackstate-agent/pkg/util/winutil

go 1.16

replace github.com/StackVista/stackstate-agent/pkg/util/log => ../log

replace github.com/StackVista/stackstate-agent/pkg/util/scrubber => ../scrubber

require (
	github.com/StackVista/stackstate-agent/pkg/util/log v0.0.0-20220817145424-3be3d1923f93 // indirect
	github.com/StackVista/stackstate-agent/pkg/util/scrubber v0.0.0-20220817145424-3be3d1923f93 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007
)
