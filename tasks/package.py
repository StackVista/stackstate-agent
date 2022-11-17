import glob
import os

from invoke import task
from invoke.exceptions import Exit

from .libs.common.color import color_message


def get_package_path(glob_pattern):
    package_paths = glob.glob(glob_pattern)
    if len(package_paths) > 1:
        raise Exit(code=1,
                   message=color_message("Too many files matching {}: {}", "red".format(glob_pattern,
                                                                                        package_paths)))  # [sts] refactored to be compatible with py2
    elif len(package_paths) == 0:
        raise Exit(code=1, message=color_message("Couldn't find any file matching {}".format(glob_pattern),
                                                 "red"))  # [sts] refactored to be compatible with py2

    return package_paths[0]


@task
def compare_size(_, new_package, stable_package, package_type, last_stable, threshold):
    mb = 1000000

    new_package_size = os.path.getsize(get_package_path(new_package))
    stable_package_size = os.path.getsize(get_package_path(stable_package))

    threshold = int(threshold)

    diff = new_package_size - stable_package_size

    # For printing purposes
    new_package_size_mb = new_package_size / mb
    stable_package_size_mb = stable_package_size / mb
    threshold_mb = threshold / mb
    diff_mb = diff / mb

    if diff > threshold:
        print(
            color_message(
                # [sts] refactored to be compatible with py2
                """{} size increase is too large:
  New package size is {:.2f}}MB
  Stable package ({}) size is {:.2f}MB
  Diff is {:.2f}MB > {:.2f}MB (max allowed diff)""".format(package_type, new_package_size_mb, last_stable,
                                                           stable_package_size_mb, diff_mb, threshold_mb),
                "red",
            )
        )
        raise Exit(code=1)

    print(
        # [sts] refactored to be compatible with py2
        """{} size increase is OK:
  New package size is {:.2f}MB
  Stable package ({}) size is {:.2f}MB
  Diff is {:.2f}MB (max allowed diff: {:.2f}MB)""".format(package_type, new_package_size_mb, last_stable,
                                                          stable_package_size_mb, diff_mb, threshold_mb)
    )
