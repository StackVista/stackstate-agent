# Keeping agent synchronized with DataDog upstream

Chose a datadog tag x.y.z from https://github.com/DataDog/datadog-agent,
ensure that the same exists tag on https://github.com/DataDog/omnibus-software.

```shell
git checkout -b upstream-x-y-z
git pull --tags https://github.com/DataDog/datadog-agent.git
git merge x.y.z

... (resolve conflicts, good luck ;) )

git commit
git push
```
