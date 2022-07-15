import logging
import pytest
from stscliv1 import CLIv1

USE_CACHE = False


@pytest.fixture
def cliv1(host, caplog) -> CLIv1:
    caplog.set_level(logging.INFO)
    return CLIv1(host, log=logging.getLogger("CLIv1"), cache_enabled=USE_CACHE)
