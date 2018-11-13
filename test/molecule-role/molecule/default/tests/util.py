import time

def wait_until(somepredicate, timeout, period=0.25, *args, **kwargs):
  mustend = time.time() + timeout
  while True:
    try:
        somepredicate(*args, **kwargs)
        return
    except:
        if time.time() >= mustend:
            print("Waiting timed out after %d" % timeout)
            raise
        time.sleep(period)
