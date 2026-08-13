# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The StackState Agent is a monitoring and observability agent forked from Datadog Agent, written primarily in Go. It collects metrics, traces, logs, and topology data from systems and applications, forwarding them to the StackState platform.

**Key StackState Differences:**
- Main configuration file is `stackstate.yaml` (not `datadog.yaml`)
- StackState-specific components in `/comp/stackstate/`
- Custom integrations repository: https://github.com/StackVista/stackstate-agent-integrations
- StackState branding and telemetry endpoints
- **Omnibus → Bazel migration (STAC-24773):** see [docs/dev/stac-24773-bazel-migration.md](docs/dev/stac-24773-bazel-migration.md) for handoff, patterns, and what not to restore after upstream merges

## Common Development Commands

### Initial Setup
```bash
# 1. Install Python dependencies (use virtual environment recommended)
pip install -r requirements.txt

# 2. Install Go tools
invoke install-tools

# 3. Install Go dependencies
invoke deps

# 4. Create dev configuration (stackstate.yaml, not datadog.yaml)
echo "api_key: <API_KEY>" > dev/dist/stackstate.yaml
```

### Building
```bash
# Build main agent (most common)
invoke agent.build --build-exclude=systemd

# Build with Python version selection
invoke agent.build --python-runtimes 2,3

# Build other components
invoke cluster-agent.build
invoke dogstatsd.build
invoke trace-agent.build
invoke process-agent.build
invoke system-probe.build
```

**Important:** When building, `--build-exclude=systemd` is commonly used in development. The agent binary is written to `bin/agent/agent` and config files are copied from `dev/dist/` to `bin/agent/dist/`.

### Testing
```bash
# Test specific package (most common during development)
invoke test --targets=./pkg/aggregator

# Skip linters during development for faster iteration
invoke test --targets=./pkg/aggregator --skip-linters

# Run tests that depend on rtloader
invoke rtloader.make && invoke rtloader.install
invoke test --targets=./pkg/collector/python
```

### Linting
```bash
# Run Go linters (do this before committing)
invoke lint-go

# Run linters on specific module/targets
invoke linter.go --targets=./pkg/collector/check,./pkg/aggregator
```

### Running the Agent
```bash
# Run the built agent
./bin/agent/agent run -c bin/agent/dist/stackstate.yaml
```

## Architecture

### Code Organization
- `/cmd/` - Entry points for agent binaries (agent, cluster-agent, dogstatsd, trace-agent, process-agent, system-probe, security-agent, etc.)
- `/pkg/` - Core Go packages (aggregator, collector, config, logs, metrics, network, trace, etc.)
- `/comp/` - Component-based architecture modules (core, logs, trace, metadata, **stackstate**)
- `/tasks/` - Python invoke task definitions for build/test/deploy automation
- `/rtloader/` - Runtime loader for Python checks

### Build Tags System
The agent uses Go build tags extensively to include/exclude features. Key tags are defined in `tasks/build_tags.py`:

**Default agent tags:** consul, containerd, cri, docker, ec2, etcd, jmx, kubeapiserver, kubelet, orchestrator, python, systemd, zk, and more.

**Component-specific tags:**
- **cluster-agent:** clusterchecks, kubeapiserver, orchestrator
- **dogstatsd:** containerd, docker, kubelet
- **trace-agent:** docker, containerd, otlp, kubeapiserver, kubelet
- **process-agent:** containerd, cri, docker, kubelet
- **system-probe:** linux_bpf, npm, pcap

When configuring your IDE (VS Code, IntelliJ), use the tags visible in the build output to enable proper Go tooling support.

### Configuration System
- Main config: `stackstate.yaml` (StackState) or `datadog.yaml` (upstream compatibility)
- Check configs: `conf.d/<check_name>.d/conf.yaml`
- Environment variable overrides: `DD_` prefix (inherited from Datadog)
- Development: Files in `dev/dist/` are copied to `bin/agent/dist/` during build

### Check System
Checks are Python or Go modules that collect metrics/topology:
- Python checks loaded via rtloader
- Go checks compiled directly into the agent
- Check configs in `conf.d/`
- Autodiscovery via Kubernetes annotations/labels

## Development Workflow

### Invoke Task System
All build/test/lint operations use Python Invoke framework. Common task categories:
- `agent.*` - Main agent operations
- `cluster-agent.*` - Cluster agent operations
- `test` - Testing tasks
- `linter.*` / `lint-*` - Linting tasks
- `rtloader.*` - Python runtime loader tasks

List all tasks: `invoke --list`

### Platform-Specific Notes
- **Linux:** Full support, all features available
- **Windows:** Full support, uses `wmi` build tag automatically
- **macOS:** Supported, some container features disabled (docker, containerd, cri, crio)
- **Tags excluded on non-Linux:** netcgo, systemd, jetson, linux_bpf, nvml, pcap, podman, trivy

### Upstream Datadog Merges
This fork is periodically merged with upstream Datadog Agent releases. This is an intensive, infrequent task with its own workflows, branding scripts, and CI patterns. See [UPSTREAM_MERGE.md](UPSTREAM_MERGE.md) for the full guide.

### Testing Cluster-Agent Helm Chart
When modifying the stackstate-agent helm chart:
1. Add test helm repo: `helm repo add stackstate-test https://helm-test.stackstate.io && helm repo update`
2. Install test chart version: `helm upgrade --install ... stackstate-test/stackstate-agent --version <version>`

## Important Build Behavior

### CMake and rtloader
If you encounter CMake errors or built an older version:
```bash
rm rtloader/CMakeCache.txt
invoke rtloader.clean
```

### Build Tag Conflicts
The build system automatically filters platform-incompatible tags. Unknown tags trigger warnings but are filtered out.

## StackState-Specific Components

- `/comp/stackstate/batcher` - StackState-specific batching logic
- `/comp/stackstate/checkmanager` - StackState check management
- `/comp/stackstate/transactionalclient` - StackState transactional client
- Integration repository separate: https://github.com/StackVista/stackstate-agent-integrations

## Notes for AI Assistants

- Always prefer dedicated tools over bash commands: `Read` over `cat`, `Edit` over `sed`, `Grep` over `grep`, `Glob` over `find`
- Configuration file is `stackstate.yaml` not `datadog.yaml` (though both may exist for compatibility)
- When referencing code, use `file_path:line_number` format
- Build tags are critical - check `tasks/build_tags.py` for component-specific requirements
- The codebase has both StackState-specific code and upstream Datadog code
