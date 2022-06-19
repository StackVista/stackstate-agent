
testinfra_hosts = ["local"]


def test_something(cliv1):
    json = cliv1("label IN ('stackpack:aws-v2')")
    assert len(json["result"]) > 500

