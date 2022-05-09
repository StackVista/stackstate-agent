module github.com/StackVista/stackstate-agent

go 1.13

// Fix tooling version
replace (
	github.com/benesch/cgosymbolizer => github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/fzipp/gocyclo => github.com/fzipp/gocyclo v0.0.0-20150627053110-6acd4345c835 // indirect
	github.com/golangci/golangci-lint => github.com/golangci/golangci-lint v1.23.1
	github.com/gordonklaus/ineffassign => github.com/gordonklaus/ineffassign v0.0.0-20200309095847-7953dde2c7bf // indirect
	// next line until pr https://github.com/ianlancetaylor/cgosymbolizer/pull/8 is merged
	github.com/ianlancetaylor/cgosymbolizer => github.com/ianlancetaylor/cgosymbolizer v0.0.0-20170921033129-f5072df9c550
	github.com/shuLhan/go-bindata => github.com/shuLhan/go-bindata v3.4.0+incompatible // indirect
)

// Internal deps fix version
replace (
	bitbucket.org/ww/goautoneg => github.com/munnerz/goautoneg v0.0.0-20120707110453-a547fc61f48d
	github.com/cihub/seelog => github.com/cihub/seelog v0.0.0-20151216151435-d2c6e5aa9fbf // v2.6
	github.com/containerd/cgroups => github.com/containerd/cgroups v0.0.0-20200327175542-b44481373989
	github.com/containerd/containerd => github.com/containerd/containerd v1.2.13
	github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20180202092358-40e2722dffea
	github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190104202606-0ac367fd6bee+incompatible
	github.com/florianl/go-conntrack => github.com/florianl/go-conntrack v0.1.1-0.20191002182014-06743d3a59db
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/iovisor/gobpf => github.com/DataDog/gobpf v0.0.0-20200131184214-6763fd92fd3f
	github.com/lxn/walk => github.com/lxn/walk v0.0.0-20180521183810-02935bac0ab8
	github.com/mholt/archiver => github.com/mholt/archiver v2.0.1-0.20171012052341-26cf5bb32d07+incompatible
	github.com/spf13/viper => github.com/DataDog/viper v1.7.1
	github.com/ugorji/go => github.com/ugorji/go v1.1.7
)

// pinned to grpc v1.26.0
replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.12.2
	google.golang.org/grpc => github.com/grpc/grpc-go v1.28.0
)

require (
	code.cloudfoundry.org/bbs v0.0.0-20200403215808-d7bc971db0db
	code.cloudfoundry.org/cfhttp/v2 v2.0.0 // indirect
	code.cloudfoundry.org/clock v1.0.0 // indirect
	code.cloudfoundry.org/consuladapter v0.0.0-20200131002136-ac1daf48ba97 // indirect
	code.cloudfoundry.org/diego-logging-client v0.0.0-20200130234554-60ef08820a45 // indirect
	code.cloudfoundry.org/executor v0.0.0-20200218194701-024d0bdd52d4 // indirect
	code.cloudfoundry.org/garden v0.0.0-20200224155059-061eda450ad9
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/lager v2.0.0+incompatible
	code.cloudfoundry.org/locket v0.0.0-20200131001124-67fd0a0fdf2d // indirect
	code.cloudfoundry.org/rep v0.0.0-20200325195957-1404b978e31e // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20180905210152-236a6d29298a // indirect
	code.cloudfoundry.org/tlsconfig v0.0.0-20200131000646-bbe0f8da39b3 // indirect
	github.com/DataDog/agent-payload v0.0.0-20200624194755-bbcbef3bd83d // 4.36.0
	github.com/DataDog/datadog-go v4.4.0+incompatible
	github.com/DataDog/datadog-operator v0.5.0-rc.2.0.20210402083916-25ba9a22e67a // indirect
	github.com/DataDog/gohai v0.0.0-20200605003749-e17d616e422a
	github.com/DataDog/gopsutil v0.0.0-20200624212600-1b53412ef321
	github.com/DataDog/mmh3 v0.0.0-20200316233529-f5b682d8c981 // indirect
	github.com/DataDog/watermarkpodautoscaler v0.3.1-logs-attributes.2.0.20211014120627-6d6a5c559fc9 // indirect
	github.com/DataDog/zstd v0.0.0-20160706220725-2bf71ec48360
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Microsoft/go-winio v0.4.16
	github.com/aws/aws-sdk-go v1.35.24
	github.com/beevik/ntp v0.3.0
	github.com/benesch/cgosymbolizer v0.0.0
	github.com/bhmj/jsonslice v0.0.0-20200323023432-92c3edaad8e2
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/clbanning/mxj v1.8.4
	github.com/containerd/cgroups v0.0.0-20200531161412-0dbf7f05ba59
	github.com/containerd/containerd v1.4.4
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c // indirect
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00 // indirect
	github.com/containerd/typeurl v1.0.1
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd v0.0.0-20190620071333-e64a0ec8b42a
	github.com/docker/docker v20.10.2+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/go-libaudit v0.4.0
	github.com/fatih/color v1.9.0
	github.com/florianl/go-conntrack v0.1.0
	github.com/go-ini/ini v1.55.0
	github.com/go-ole/go-ole v1.2.4
	github.com/go-test/deep v1.0.5 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/gopacket v1.1.17
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/consul/api v1.4.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-00010101000000-000000000000 // indirect
	github.com/iovisor/gobpf v0.0.0-20200329161226-8b2cce9dac28
	github.com/itchyny/gojq v0.10.2
	github.com/json-iterator/go v1.1.12
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/kubernetes-sigs/custom-metrics-apiserver v0.0.0-20210311094424-0ca2b1909cdc
	github.com/lib/pq v1.10.5 // indirect
	github.com/lxn/walk v0.0.0-20191128110447-55ccb3a9f5c1
	github.com/lxn/win v0.0.0-20191128105842-2da648fda5b4
	github.com/mdlayher/netlink v1.1.0
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/miekg/dns v1.1.35
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pierrec/lz4 v2.5.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/shirou/gopsutil v2.20.3+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4
	github.com/soniah/gosnmp v1.26.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	github.com/tinylib/msgp v1.1.2
	github.com/twmb/murmur3 v1.1.3
	github.com/urfave/negroni v1.0.0
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae
	github.com/vito/go-sse v1.0.0 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/mobile v0.0.0-20201217150744-e6ae53a27f4f
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9
	golang.org/x/text v0.3.6
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.org/x/tools v0.1.4 // indirect
	gomodules.xyz/jsonpatch/v3 v3.0.1
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa
	google.golang.org/grpc v1.42.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.29.0-rc.1.0.20210226170446-a8dc39ec3484
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/zorkian/go-datadog-api.v2 v2.29.0
	k8s.io/api v0.21.5
	k8s.io/apimachinery v0.21.5
	k8s.io/apiserver v0.21.5
	k8s.io/autoscaler/vertical-pod-autoscaler v0.9.2
	k8s.io/client-go v0.21.5
	k8s.io/cri-api v0.21.5
	k8s.io/klog v1.0.1-0.20200310124935-4ad0115ba9e4 // indirect; Min version that includes fix for Windows Nano
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/kube-state-metrics v1.9.7
	k8s.io/metrics v0.21.5
)

// Pinned so it includes fix for Windows Nano
replace k8s.io/klog => k8s.io/klog v1.0.1-0.20200310124935-4ad0115ba9e4

replace (
	k8s.io/api => k8s.io/api v0.21.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.5
	k8s.io/apiserver => k8s.io/apiserver v0.21.5
	k8s.io/autoscaler => k8s.io/autoscaler v0.0.0-20191115143342-4cf961056038
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.5
	k8s.io/code-generator => k8s.io/code-generator v0.21.5
	k8s.io/component-base => k8s.io/component-base v0.21.5
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.5
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.5
	k8s.io/cri-api => k8s.io/cri-api v0.21.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.5
	k8s.io/kube-state-metrics => k8s.io/kube-state-metrics v1.9.6-0.20200413182837-dbbe062e36a4
	k8s.io/kubectl => k8s.io/kubectl v0.21.5
	k8s.io/kubelet => k8s.io/kubelet v0.21.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.5
	k8s.io/metrics => k8s.io/metrics v0.21.5
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.5
)

exclude github.com/containerd/containerd v1.5.0-beta.1
