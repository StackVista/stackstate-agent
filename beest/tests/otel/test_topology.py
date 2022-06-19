import util

testinfra_hosts = ["local"]


def test_not_empty(cliv1):
    def assert_it():
        json = cliv1("label IN ('stackpack:aws-v2')")
        assert len(json["result"]) > 500

    util.wait_until(assert_it, 30, 5)
