[![Go Report Card](https://goreportcard.com/badge/github.com/disaster37/check_opensearch)](https://goreportcard.com/report/github.com/disaster37/check_opensearch)
[![GoDoc](https://godoc.org/github.com/disaster37/check_opensearch?status.svg)](http://godoc.org/github.com/disaster37/check_opensearch)
[![codecov](https://codecov.io/gh/disaster37/check_opensearch/branch/2.x/graph/badge.svg)](https://codecov.io/gh/disaster37/check_opensearch/branch/2.x)


# check_opensearch
Nagios plugin to check the healtch of Opensearch cluster

## Contribute

You PR are always welcome. Please use the righ branch to do PR:
 - 2.x for Opensearch 2.x
Don't forget to add test if you add some functionalities.

To build, you can use the following command line:
```sh
go build
```

To lauch golang test, you can use the folowing command line:
```bash
docker-compose up -d
go test -v ./...
```

## CLI

### Global options

The following parameters are available for all commands line :
- **--url**: The Opensearch URL. For exemple https://opensearch.company.com. Alternatively you can use environment variable `OPENSEARCH_URL`.
- **--user**: The login to connect on Opensearch. Alternatively you can use environment variable `OPENSEARCH_USER`.
- **--password**: The password to connect on Opensearch. Alternatively you can use environment variable `OPENSEARCH_PASSWORD`.
- **--self-signed-certificate**: Disable the check of server SSL certificate
- **--debug**: Enable the debug mode
- **--help**: Display help for the current command


You can set also this parameters on yaml file (one or all) and use the parameters `--config` with the path of your Yaml file.
```yaml
---
url: https://opensearch.company.com
user: admin
password: changeme
```

### Check if indice are locked by storage pressure

Command `check-indice-locked` permit to check if indice provided is not locked by storage pressure.
If you should to check all indice, you can put `_all` as indice name.

You need to set the following parameters:
- **--indice**: The indice name to check

It return the following perfdata:
- **nbIndices**: the number of indices returned
- **nbIndicesLocked**: the number of indices locked


Sample of command:
```bash
./check_opensearch --url https://localhost:9200 --user admin --password changeme check-indice-locked --indice _all
```

Response:
```bash
OK - No indice locked (6/6)|nbIndices=6;;;; nbIndicesLocked=0;;;;
```


### Check ISM errors on indice

Command `check-ism-indice` permit to check if ISM policy failed on given indice.
If you should to check all indice, you can put `_all` as indice name.

You need to set the following parameters:
- **--indice**: The indice name
- **--exclude**: (optional) The indice name you should to exclude

It return the following perfdata:
- **nbIndicesFailed**: the number of indices with ILM error

Sample of command:
```bash
./check_opensearch --url https://localhost:9200 --user admin --password changeme check-ism-indice --indice _all
```

Response:
```bash
OK - No error found on indice _all|NbIndiceFailed=0;;;; 
```


### Check if there are snapshot errors

Command `check-repository-snapshot` permit to check if there are snapshot error on given repository.

You need to set the following parameters:
- **--repository**: The repository name where you should to check snapshots

It return the following perfdata:
- **nbSnapshot**: the number of snapshot
- **nbSnapshotFailed**: the number of failed snapshot

Sample of command:
```bash
./check_opensearch --url https://localhost:9200 --user admin --password changeme check-repository-snapshot --repository snapshot
```

Response:
```bash
OK - No snapshot on repository snapshot|NbSnapshot=0;;;; NbSnapshotFailed=0;;;;
```

### Check if there are SM policies errors

Command `check-sm-policy` permit to check if there are SM policies error.

You can to set the following parameters if you should to check only one policy:
- **--name**: The policy name you should to check

It return the following perfdata:
- **nbSLMPolicy**: the number of SLM policy
- **nbSLMPolicyFailed**: the number of failed SLM policy

Sample of command:
```bash
./check_opensearch --url https://localhost:9200 --user admin --password changeme check-sm-policy
```

Response:
```bash
OK - All SLM policies are ok (1/1)|NbSLMPolicy=1;;;; NbSLMPolicyFailed=0;;;;
```

### Check Transform errors

Command `check-transform` permit to check if tranform failed.
If you should to check all tranform, you can put `_all` as transform name.

You need to set the following parameters:
- **--name**: The transform name
- **--exclude**: (optional) The transform name you should to exclude

It return the following perfdata:
- **nbTransformFailed**: the number of transform failed
- **nbTransformStarted**: the number of transform started
- **nbTransformStopped**: the number of transform stopped

Sample of command:
```bash
./check_opensearch --url https://localhost:9200 --user admin --password changeme check-transform --name _all
```

Response:
```bash
OK - No error found on indice _all|NbIndiceFailed=0;;;; 
```