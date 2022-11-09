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

# Upgrade submodules interdependencies

```shell
# setup new versions
inv -e release.update-modules 2.19.0[-rc.4]
# make commit
git commit
# tag new versions
inv -e release.tag-version 2.19.0-rc.4 [--no-push (to check what will happen, will create local tags though)] [--force (to overwrite previous run with --no-push)]
# check if everything settled correclty - no changes (to go.mod) after this command
inv -e deps
```
