FROM artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-runner-gitlab:deb7_20211210

COPY scripts/entry_point.sh /entry_point.sh
COPY scripts/shell.sh /shell.sh

RUN chmod +x /entry_point.sh
RUN chmod +x /shell.sh

ENTRYPOINT ["/entry_point.sh"]
