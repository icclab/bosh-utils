# Get-CredHub-Var

A small cmd line program that lets you very easily search and retrieve any secret stored on CredHub, including certificates.
Search results are sorted and categorized by the BOSH deployment the variable is in. If a single password is found then it will be copied to your clipboard automatically. It also provides the ability to backup all CredHub vars into a yaml file that can be directly imported by CredHub.

## Install

To build simply run `go build` and then put the binary in your path, for example on Linux you can do `mv get-credhub-var /usr/local/bin/`. If you are using Linux, then
you can simply choose a release and download the binary and similarly move to path by `mv get-credhub-var-Linux64 /usr/local/bin/get-credhub-var`.

## Usage

To run you need the following environment variables specified correctly for the CredHub deployment you want to target:

- `BOSH_ENVIRONMENT`: Should look something like this `https://192.168.50.6:25555`
- `CREDHUB_CLIENT`
- `CREDHUB_SECRET`
- `CREDHUB_CA_CERT`

The image below shows the result of running `get-credhub-var ""`, which retrieves all the variables under all deployments.
The `/` section contains variables that **not** part of any BOSH deployment.

![Example results](imgs/example.png)

Since BOSH secrets are structured like this: `/<BOSH-env-name>/<deployment-name>/<secret-name>`, to get all the variables in a deployment you can run with
`get-credhub-var /<deployment-name>/` (keeping the slashes).

You can also run with `get-credhub-var my-search-term -v` to show the value of any secrets found. `-V` can be used if only one variable is found to print all the details
about that secret.

### Backup

running `get-credhub-var backup` will backup all your CredHub secrets into a file called `bosh-backup.yml`.