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
	github.com/docker/distribution => github.com/docker/distribution v2.8.0+incompatible
	github.com/florianl/go-conntrack => github.com/florianl/go-conntrack v0.1.1-0.20191002182014-06743d3a59db
	github.com/iovisor/gobpf => github.com/DataDog/gobpf v0.0.0-20200131184214-6763fd92fd3f
	github.com/lxn/walk => github.com/lxn/walk v0.0.0-20180521183810-02935bac0ab8
	github.com/mholt/archiver => github.com/mholt/archiver v2.0.1-0.20171012052341-26cf5bb32d07+incompatible
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	github.com/spf13/viper => github.com/DataDog/viper v1.7.1
	github.com/ugorji/go => github.com/ugorji/go v1.1.7
)

// pinned to grpc v1.26.0
replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.12.2
	google.golang.org/grpc => github.com/grpc/grpc-go v1.26.0
)

require (
	bitbucket.org/ww/goautoneg v0.0.0-20120707110453-75cd24fc2f2c // indirect
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
	github.com/DATA-DOG/go-sqlmock v1.4.1 // indirect
	github.com/DataDog/agent-payload v0.0.0-20200624194755-bbcbef3bd83d // 4.36.0
	github.com/DataDog/datadog-go v3.5.0+incompatible
	github.com/DataDog/datadog-operator v0.2.1-0.20200527110245-7850164045c8
	github.com/DataDog/gohai v0.0.0-20200605003749-e17d616e422a
	github.com/DataDog/gopsutil v0.0.0-20200624212600-1b53412ef321
	github.com/DataDog/mmh3 v0.0.0-20200316233529-f5b682d8c981 // indirect
	github.com/DataDog/watermarkpodautoscaler v0.1.0
	github.com/DataDog/zstd v0.0.0-20160706220725-2bf71ec48360
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5
	github.com/aws/aws-sdk-go v1.30.5
	github.com/beevik/ntp v0.3.0
	github.com/benesch/cgosymbolizer v0.0.0
	github.com/bhmj/jsonslice v0.0.0-20200323023432-92c3edaad8e2
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/clbanning/mxj v1.8.4
	github.com/codedellemc/goscaleio v0.0.0-20170830184815-20e2ce2cf885 // indirect
	github.com/containerd/cgroups v0.0.0-20190919134610-bf292b21730f
	github.com/containerd/containerd v1.3.2
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c // indirect
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00 // indirect
	github.com/containerd/typeurl v1.0.0
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd v0.0.0-20190620071333-e64a0ec8b42a
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/coreos/prometheus-operator v0.29.0 // indirect
	github.com/d2g/dhcp4 v0.0.0-20170904100407-a1d1b6c41b1c // indirect
	github.com/d2g/dhcp4client v0.0.0-20170829104524-6e570ed0a266 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/go-libaudit v0.4.0
	github.com/elazarl/go-bindata-assetfs v1.0.0 // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6 // indirect
	github.com/fatih/color v1.9.0
	github.com/florianl/go-conntrack v0.1.0
	github.com/frankban/quicktest v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-bindata/go-bindata v3.1.2+incompatible // indirect
	github.com/go-ini/ini v1.55.0
	github.com/go-ole/go-ole v1.2.4
	github.com/go-test/deep v1.0.5 // indirect
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/gopacket v1.1.17
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/gregjones/httpcache v0.0.0-20190203031600-7a902570cb17 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hako/durafmt v0.0.0-20191009132224-3f39dc1ed9f4 // indirect
	github.com/hashicorp/consul/api v1.4.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/helm/helm-2to3 v0.5.1 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40 // indirect
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365 // indirect
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-00010101000000-000000000000 // indirect
	github.com/iovisor/gobpf v0.0.0-20200329161226-8b2cce9dac28
	github.com/itchyny/gojq v0.10.2
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/jteeuwen/go-bindata v0.0.0-20151023091102-a0ff2567cfb7 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/kr/fs v0.0.0-20131111012553-2788f0dbd169 // indirect
	github.com/lib/pq v1.2.0 // indirect
	github.com/lxn/walk v0.0.0-20191128110447-55ccb3a9f5c1
	github.com/lxn/win v0.0.0-20191128105842-2da648fda5b4
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/martinlindhe/base36 v1.0.0 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/maxbrunsfeld/counterfeiter v0.0.0-20181017030959-1aadac120687 // indirect
	github.com/mdlayher/netlink v1.1.0
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mholt/caddy v0.0.0-20180213163048-2de495001514 // indirect
	github.com/miekg/dns v1.1.27
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/prom-label-proxy v0.1.1-0.20191016113035-b8153a7f39f1 // indirect
	github.com/operator-framework/go-appr v0.0.0-20180917210448-f2aef88446f2 // indirect
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200321030439-57b580e57e88 // indirect
	github.com/operator-framework/operator-marketplace v0.0.0-20190216021216-57300a3ef3ba // indirect
	github.com/operator-framework/operator-registry v1.6.2-0.20200330184612-11867930adb5 // indirect
	github.com/operator-framework/operator-sdk v0.8.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/petar/GoLLRB v0.0.0-20130427215148-53be0d36a84c // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pierrec/lz4 v2.5.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20191025130928-9355dd04f4b3 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/shirou/gopsutil v2.20.3+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4
	github.com/sigma/go-inotify v0.0.0-20181102212354-c87b6cf5033d // indirect
	github.com/soniah/gosnmp v1.26.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stevvooe/resumable v0.0.0-20180830230917-22b14a53ba50 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	github.com/thanos-io/thanos v0.11.0 // indirect
	github.com/tinylib/msgp v1.1.2
	github.com/twmb/murmur3 v1.1.3
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/urfave/negroni v1.0.0
	github.com/vishvananda/netns v0.0.0-20171111001504-be1fbeda1936
	github.com/vito/go-sse v1.0.0 // indirect
	github.com/vmware/photon-controller-go-sdk v0.0.0-20170310013346-4a435daef6cc // indirect
	github.com/xanzy/go-cloudstack v0.0.0-20160728180336-1e2cbf647e57 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	github.com/zorkian/go-datadog-api v2.28.0+incompatible // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	go.opentelemetry.io/proto/otlp v0.11.0 // indirect
	golang.org/x/mobile v0.0.0-20190719004257-d2bd2a29d028
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1
	golang.org/x/text v0.3.6
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	gomodules.xyz/jsonpatch/v3 v3.0.1
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a
	google.golang.org/grpc v1.42.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.23.1
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	gopkg.in/gorp.v1 v1.7.2 // indirect
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/zorkian/go-datadog-api.v2 v2.29.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	k8s.io/api v0.21.5
	k8s.io/apimachinery v0.21.5
	k8s.io/apiserver v0.21.5
	k8s.io/autoscaler/vertical-pod-autoscaler v0.9.2
	k8s.io/client-go v0.21.5
	k8s.io/cri-api v0.21.5
	k8s.io/klog v1.0.1-0.20200310124935-4ad0115ba9e4 // Min version that includes fix for Windows Nano
	k8s.io/klog/v2 v2.9.0
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
	k8s.io/kube-state-metrics/v2 v2.1.1
	k8s.io/kubelet v0.21.5
	k8s.io/metrics v0.21.5
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
)

require github.com/docker/distribution v2.7.1+incompatible // indirect

// Pinned so it includes fix for Windows Nano
replace k8s.io/klog => k8s.io/klog v1.0.1-0.20200310124935-4ad0115ba9e4

// Include bug fixes not released upstream (yet)
// - https://github.com/kubernetes/kube-state-metrics/pull/1610
// - https://github.com/kubernetes/kube-state-metrics/pull/1584
replace k8s.io/kube-state-metrics/v2 => github.com/DataDog/kube-state-metrics/v2 v2.1.2-0.20211109105526-c17162ee2798

replace (
	k8s.io/api => k8s.io/api v0.21.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.5
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
	k8s.io/metrics => k8s.io/metrics v0.21.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/apiserver => k8s.io/apiserver v0.21.5
	k8s.io/client-go => k8s.io/client-go v0.21.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	k8s.io/cri-api => k8s.io/cri-api v0.21.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
	k8s.io/kubelet => k8s.io/kubelet v0.21.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
)
