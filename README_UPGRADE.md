# Keeping agent synchronized with DataDog upstream

In short, to upgrade the agent to a new datadog verison, we use an aproach of rebasing/cherry-picking our changes on top of the
version of the agent we want to upgrade to. This will ultimately yield a new base-branch called stackstate-X.Y.Z where X.Y.Z is the
datadog version we put our work on top off.

Because we have merge commits in our history, we cannot rebase/cherry-pick directly onto datadog. We first squash our work into a single commit, to get rid of merge commits.

Lucidchart here: **https://lucid.app/lucidspark/2f9240e3-dca7-4ef1-b796-e7742fc4f1e2/edit**

We use the following version (git hash) shortnames:
- `F`: The datadog version we currently base ourselves on
- `Fsts`: The git commit denoting the latest HEAD of the patch we put on datadog version `F`
- `Fsq`: the git commit contain all our squashed changed w.r.t datadog
- `T`: The datadog version we want to go to

The procedure now is as follows:
- Create a branch called `squashed-F` based off `Fsts`.
- Squash all commits up to (but excluding) commit `F` in a single commit (we call it `Fsq`). Give it a commit message along the lines of "All work up to datadog version F"
- Create a new branch called `stackstate-T` based on the datadog commit we want to move to (`T`)
- Create a new branch `Fsq-on-T` based on `stackstate-T`. We cherry `Fsq` into `Fsq-on-T`
- Create an MR `Fsts-on-T` into `stackstate-T` which can be reviewed and round up the merge (should be fast forward merge due to linear history).
- If patches come in during the agent upgrade procedure, those are cherry-picked later into `Fsts-on-T

Pro tip:
- It can be helpful to open a temporary MR from `Fsts` into `F` (to be discarded), to show the delta that we currently have with datadog for reference during the upgrade.
