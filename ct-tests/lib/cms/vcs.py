# Copyright 2020-2021 Hewlett Packard Enterprise Development LP

"""
VCS-related test helper functions for CMS tests
"""

from .k8s import get_vcs_username_password
from .helpers import debug, info, raise_test_exception_error, run_cmd_list

def clone_vcs_repo(tmpdir):
    """
    Clone the config-management vcs repo into the specified temporary directory.
    Sets user.name and user.email for that repo.
    Returns the repo directory.
    """
    try:
        vcsuser, vcspass = get_vcs_username_password()
    except Exception as e:
        raise_test_exception_error(e, "to get vcs username and password from kubernetes")
    run_cmd_list(
        ["git", "clone", "https://%s:%s@api-gw-service-nmn.local/vcs/cray/config-management.git" % (vcsuser, vcspass)],
        cwd=tmpdir)
    git_repo_dir="%s/config-management" % tmpdir
    info("Cloned vcs repo to directory %s" % git_repo_dir)
    run_cmd_list(["git", "-C", git_repo_dir, "config", "user.email", "catfood@dogfood.mil"])
    debug("Set user.email for cloned repo")
    run_cmd_list(["git", "-C", git_repo_dir, "config", "user.name", "Rear Admiral Joseph Catfood"])
    debug("Set user.name for cloned repo")
    return git_repo_dir

def create_vcs_branch(repo_dir, branchname, base_branch="master"):
    """
    In the specified repo dir:
    1) checkout the base branch
    2) create and checkout a new branch with the specified name
    3) modify the motd yaml file to include the name of the branch
    4) Commit and push the change to vcs
    """
    git_motd_yaml="%s/roles/motd/defaults/main.yml" % repo_dir
    debug("Returning to master branch")
    run_cmd_list(["git", "checkout", base_branch], cwd=repo_dir)
    debug("Creating branch %s" % branchname)
    run_cmd_list(["git", "checkout", "-b", branchname], cwd=repo_dir)
    debug("Modify %s" % git_motd_yaml)
    with open(git_motd_yaml, "rt") as f:
        all_lines = f.read().splitlines()
    with open(git_motd_yaml, "wt+") as f:
        for line in all_lines[:-1]:
            f.write("%s\n" % line)
        f.write("%s branch=%s\n" % (all_lines[-1], branchname))
    debug("Add & commit change")
    run_cmd_list(["git", "commit", "-am", "CMS tests are the best"], cwd=repo_dir)
    debug("Push commit")
    run_cmd_list(["git", "push", "--set-upstream", "origin", branchname], cwd=repo_dir)

def remove_vcs_test_branches(repo_dir, test_vcs_branches):
    """
    Delete the specified branches from vcs
    """
    blist = list(test_vcs_branches)
    for branch in blist:
        debug("Deleting branch %s" % branch)
        run_cmd_list(["git", "push", "origin", "--delete", branch], cwd=repo_dir)
        test_vcs_branches.remove(branch)

