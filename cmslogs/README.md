# cmslogs

This is a utility to collect relevant files from the node on which it is run, in order to debug a CMS test failure. It is installed to /usr/local/bin/cmslogs. It is bundled with the cmsdev tool.

## What it collects

By default, this tool collects all of the following files and directories (but accepts flags to omit them):
* /etc/cray-release
* /etc/motd
* /opt/cray/tests
* /tmp/cray/tests

In addition, it records the output of the following two commands:
* cksum /usr/local/bin/cmsdev
* rpm -qa

## Usage

```
usage: cmslogs [-h] [-f] [--no-cmsdev-sum] [--no-cray-release] [--no-motd]
               [--no-opt-cray-tests] [--no-opt-cray-tests-all] [--no-rpms]
               [--no-tmp-cray-tests]

Collect files for test debug and stores them in /tmp/cmslogs.tar

optional arguments:
  -h, --help            show this help message and exit
  -f                    Overwrite outfile (/tmp/cmslogs.tar) if it exists
  --no-cmsdev-sum       Do not record output of cksum /usr/local/bin/cmsdev
                        command
  --no-cray-release     Do not collect /etc/cray-release
  --no-motd             Do not collect /etc/motd
  --no-opt-cray-tests   Do not collect /opt/cray/tests directory (except logs)
  --no-opt-cray-tests-all
                        Do not collect /opt/cray/tests directory (including
                        logs)
  --no-rpms             Do not collect output of rpm -qa command
  --no-tmp-cray-tests   Do not collect /tmp/cray/tests directory
```

## Contributing
Pull requests are welcome.
