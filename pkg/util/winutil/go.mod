module github.com/StackVista/stackstate-agent/pkg/util/winutil

go 1.21

replace github.com/StackVista/stackstate-agent/pkg/util/log => ../log

replace github.com/StackVista/stackstate-agent/pkg/util/scrubber => ../scrubber

require (
	github.com/StackVista/stackstate-agent/pkg/util/log v0.19.0-rc.4 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007
)
