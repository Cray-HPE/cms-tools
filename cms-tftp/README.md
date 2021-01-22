# CRAY-TFTP-UPLOAD Tool

`cray-tftp-upload` is a utility for uploading files to be served by the *cray-tftp* service. The tool executes a kubectl copy to copy the file to the storage location shared by *cray-ipxe* and *cray-tftp*.  Since the *cray-ipxe* service is the entity that mounts the shared storage with write permission, the `cray-tftp-upload` tool connects to the *cray-ipxe* service to perform the copy.

## Execution

To upload a file to *cray-tftp*, perform the following command from `ncn-w001`:

`cray-tftp-upload path_to_your_file`

## Example output:

1. Failure to provide a file to upload:
```
# cray-tftp-upload
No file given to upload. Exiting.
Usage:
cray-tftp-upload: <upload-file>
```

2. File "foo" doesn't exist:
```
# cray-tftp-upload foo
foo is not a file. Exiting.
```

3. Uploading a file named "test":
```
# cray-tftp-upload test
Uploading file: test
Defaulting container name to cray-ipxe.
Successfully uploaded test!
#
```