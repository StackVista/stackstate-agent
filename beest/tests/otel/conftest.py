import logging
import pytest
from stscliv1 import CLIv1

USE_CACHE = False


@pytest.fixture
def cliv1(host, caplog) -> CLIv1:
    caplog.set_level(logging.INFO)
    return CLIv1(host, log=logging.getLogger("CLIv1"), cache_enabled=USE_CACHE)


@pytest.fixture
def ansible_var(host):
    def _retrieve_var(name):
        raw_vars = host.ansible.get_variables()
        if name in raw_vars and (type(raw_vars[name]) != str or "{{" not in raw_vars[name]):
            # No interpolation needed, we return the raw value
            return raw_vars[name]
        else:
            # This allows variable interpolation
            # https://stackoverflow.com/questions/57820998/accessing-ansible-variables-in-molecule-test-testinfra
            return host.ansible("debug", "msg={{ " + name + " }}")["msg"]
    return _retrieve_var