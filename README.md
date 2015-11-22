# azkube-kvbs

## Overview

kbvs = "KeyVault Bootstrap"

KeyVault will drop a certificate and a PKCS#8 private key onto our Azure Compute instances when we create the cluster or scale it up.

`azkube-kbvs` uses those to download the relevant secrets for this machines role
