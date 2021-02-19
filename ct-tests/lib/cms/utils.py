# Copyright 2020-2021 Hewlett Packard Enterprise Development LP

"""
CMS test helper functions that
- span multiple services
or
- don't neatly fit into one of the other categories
"""

from .bss import bss_host_nid, bss_host_xname, get_bss_compute_nodes
from .helpers import CMSTestError, debug, error, info, raise_test_error, \
                     run_cmd_list, is_pingable, run_command_via_ssh, \
                     run_scp, ssh_command_passes
from .hsm import get_hsm_xname_list
from .k8s import get_csm_private_key
import os
import stat
import tempfile

def get_bss_hsm_compute_nodes(use_api):
    """
    Return BSS host entries for all compute nodes that are Enabled in BSS
    and listed in HSM with Status Populated
    """
    hsm_node_xnames = get_hsm_xname_list(use_api)
    debug("Found the following xnames for Populated nodes in hsm: %s" % str(hsm_node_xnames))
    if not hsm_node_xnames:
        raise_test_error("No nodes found in HSM inventory with Status Populated")
    bss_compute_nodes = get_bss_compute_nodes(use_api)
    debug("Found the following Enabled Compute nodes in BSS: %s" % str(bss_compute_nodes))
    filtered_list = [ host for host in bss_compute_nodes if bss_host_xname(host) in hsm_node_xnames ]
    debug("So the intersection is: %s" % str(filtered_list))
    if not filtered_list:
        raise_test_error("No compute nodes found that both have Status Populated in HSM and are Enabled in BSS")
    return filtered_list

def get_compute_nids_xnames(use_api, nids=None, xnames=None, groups=None, min_required=0):
    """
    First, from BSS and HSM, generate a list of compute nodes.
    From that, generate a map between compute node NIDs and xnames.
    
    If no NIDs, xnames, or groups have been specified to the function,
    then using capmc, remove from this mapping any nodes which do not
    report a power state of off or on.
    
    If any NIDs, xnames, or groups have been specified, then remove from
    the mapping any nodes which do not match one of the parameters passed into
    the function.
    
    Verify that the resulting mapping contains at least the minimum required
    number of nodes. Then return nid -> xname and xname -> nid mappings.
    """
    bss_hsm_compute_nodes = get_bss_hsm_compute_nodes(use_api)
    bss_hsm_nids_to_xnames = { bss_host_nid(host): bss_host_xname(host) for host in bss_hsm_compute_nodes }
    if nids or xnames or groups:
        if xnames:
            xnames_set = set(xnames)
        else:
            xnames_set = set()
        if groups:
            for gname in groups:
                info("Listing xnames in HSM group %s" % gname)
                xnames_from_gname = list_hsm_group_members(use_api=use_api, group_name=gname)
                info("HSM group %s contains xnames %s" % (gname, ', '.join(xnames_from_gname)))
                xnames_set.update(xnames_from_gname)
        if nids:
            nids_set = set(nids)
        else:
            nids_set = set()
        nid_to_xname = { n:x for (n, x) in bss_hsm_nids_to_xnames.items() if n in nids_set or x in xnames_set }
        omitted_nids = [ str(n) for n in nids_set if n not in nid_to_xname.keys() ]
        omitted_xnames = [ x for x in xnames_set if x not in nid_to_xname.values() ]
        errors=False
        if omitted_nids:
            error("One or more NIDs not found Enabled in BSS or Populated in HSM: %s" % ", ".join(omitted_nids))
            errors=True
        if omitted_xnames:
            error("One or more xnames not found Enabled in BSS or Populated in HSM: %s" % ", ".join(omitted_xnames))
            errors=True
        if errors:
            raise_test_error("One or more NIDs and/or xnames not found Enabled in BSS or Populated in HSM")
    else:
        nid_to_capmc_status = get_capmc_node_status(use_api=use_api, 
                                                    nids=list(bss_hsm_nids_to_xnames.keys()),
                                                    return_undefined=True)
        on_off_nids = { n for n,p in nid_to_capmc_status.items() if p in { "off", "on" } }
        nid_to_xname = { n: x for (n, x) in bss_hsm_nids_to_xnames.items() if n in on_off_nids }
        debug("Here is the mapping for NIDs with power states on or off: %s" % str(nid_to_xname))
    if len(nid_to_xname) < min_required:
        raise_test_error("%d compute node(s) found, but this test requires at least %d" % (len(nid_to_xname), min_required))
    xname_to_nid = { x: n for (n, x) in nid_to_xname.items() }
    return nid_to_xname, xname_to_nid

def node_hostname(xname):
    """
    Return the hostname for the specified xname
    """
    return xname
    
def validate_node_hostname(nid, xname):
    """
    Validate that the specified node has a resolvable hostname
    """
    nh = node_hostname(xname)
    try:
        run_cmd_list(["host", nh])
    except CMSTestError:
        error("Unable to resolve hostname (%s) for nid %d" % (nh, nid))
        raise

def validate_node_hostnames(nid_to_xname):
    """
    Verify that all of our target nodes have resolvable hostnames
    """
    for nid, xname in nid_to_xname.items():
        validate_node_hostname(nid, xname)

def is_xname_pingable(xname):
    """
    Returns True if node is pingable, false otherwise
    """
    return is_pingable(node_hostname(xname))

def csm_key_tmpfile(dir=None):
    """
    Helper function for the following ssh/scp functions.
    Returns a NamedTemporaryFile which they can use to write the CSM private key data
    """
    csm_key = get_csm_private_key()
    f = tempfile.NamedTemporaryFile(mode="wt", encoding="ascii", dir=dir, prefix="csm-key-", 
                                    suffix=".tmp", delete=True)
    debug("Writing CSM private key to temporary file %s" % f.name)
    f.write("%s\n" % csm_key)
    f.flush()
    debug("Setting 600 permissions on temporary file %s" % f.name)
    try:
        os.chmod(f.name, stat.S_IRUSR|stat.S_IWUSR)
    except (FileNotFoundError, PermissionError) as e:
        raise CMSTestError("Unable to set file permissions on %s" % f.name, log_error=False) from e
    return f

def run_command_on_xname_via_ssh(xname, cmdstring, use_csm_key=True, tmpdir=None, **kwargs):
    """
    Determines the hostname for the specified xname, then
    runs the specified command via ssh on it.
    Returns True if this succeeds (both the ssh and the command), False otherwise.
    """
    if use_csm_key:
        with csm_key_tmpfile(dir=tmpdir) as f:
            return run_command_on_xname_via_ssh(xname=xname, cmdstring=cmdstring, use_csm_key=False, 
                                                identity_file=f.name, **kwargs)
    return run_command_via_ssh(node_hostname(xname), cmdstring, **kwargs)

def ssh_command_passes_on_xname(xname, cmdstring, use_csm_key=True, tmpdir=None, **kwargs):
    """
    Determines the hostname for the specified xname, then
    runs the specified command via ssh on it.
    Returns True if this succeeds (both the ssh and the command), False otherwise.
    """
    if use_csm_key:
        with csm_key_tmpfile(dir=tmpdir) as f:
            return ssh_command_passes_on_xname(xname=xname, cmdstring=cmdstring, use_csm_key=False, 
                                               identity_file=f.name, **kwargs)
    return ssh_command_passes(node_hostname(xname), cmdstring, **kwargs)

def scp_to_xname(local_file, xname, use_csm_key=True, tmpdir=None, **kwargs):
    """
    Determines the hostname for the specified xname, then calls run_scp using it
    """
    if use_csm_key:
        with csm_key_tmpfile(dir=tmpdir) as f:
            scp_to_xname(local_file=local_file, xname=xname, use_csm_key=False, 
                         identity_file=f.name, **kwargs)
        return
    run_scp(local_file=local_file, target_host=node_hostname(xname), **kwargs)
