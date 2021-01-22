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

# CRAY-UPLOAD-RECOVERY-IMAGES

`cray-upload-recovery-images` is a utility for uploading the cray BMC recovery files to be served by the *cray-tftp* service. The tool uses the cray cli (*fas*, *artifacts*) and *cray-tftp* to download the s3 recovery images (as remembered by FAS) then upload them into the PVC that is used by *cray-tfpt*.

## Execution

To upload the recovery images file to *cray-tftp*, perform the following command from `ncn-w001`:

`cray-upload-recovery-images`

## Example output:

1. Run the command

	```
	# cray-upload-recovery-images
	Attempting to retrieve ChassisBMC .itb file
	s3:/fw-update/d7bb5be9eecc11eab18c26c5771395a4/cc-1.3.10.itb
	d7bb5be9eecc11eab18c26c5771395a4/cc-1.3.10.itb
	
	Uploading file: /tmp/cc.itb
	Defaulting container name to cray-ipxe.
	Successfully uploaded /tmp/cc.itb!
	removed /tmp/cc.itb
	ChassisBMC recovery image upload complete
	========================================
	Attempting to retrieve NodeBMC .itb file
	s3:/fw-update/d81157f7eecc11ea943d26c5771395a4/nc-1.3.10.itb
	d81157f7eecc11ea943d26c5771395a4/nc-1.3.10.itb
	
	Uploading file: /tmp/nc.itb
	Defaulting container name to cray-ipxe.
	Successfully uploaded /tmp/nc.itb!
	removed /tmp/nc.itb
	NodeBMC recovery image upload complete
	========================================
	Attempting to retrieve RouterBMC .itb file
	s3:/fw-update/d85398f2eecc11ea94ff26c5771395a4/rec-1.3.10.itb
	d85398f2eecc11ea94ff26c5771395a4/rec-1.3.10.itb
	
	Uploading file: /tmp/rec.itb
	Defaulting container name to cray-ipxe.
	Successfully uploaded /tmp/rec.itb!
	removed /tmp/rec.itb
	RouterBMC recovery image upload complete
	```

