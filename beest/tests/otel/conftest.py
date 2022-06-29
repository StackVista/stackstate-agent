import pytest
from stscliv1 import CLIv1

@pytest.fixture
def cliv1(host) -> CLIv1:
    return CLIv1(host)
