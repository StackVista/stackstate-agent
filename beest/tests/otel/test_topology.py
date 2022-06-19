
testinfra_hosts = ["local"]


def test_something(cli):
    json = cli("label IN ('stackpack:aws-v2')")
    assert len(json["result"]) > 500

