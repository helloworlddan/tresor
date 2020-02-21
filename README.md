# Tresor

Tresor is an asymmetric client-side encryption frontend for Google Cloud Storage using OpenPGP

## Setup

Tresor uses a configuration file at `~/.tresor.yaml`. It looks like this:

```yaml
bucket: gcs-bucket-name
public_key: /path/to/armored/public/key.asc
private_key: /path/to/armored/public/key.asc
ascii_armor: true # Armored objects?
object_signing: false # Signed objects?
```

Create this file and configure your environment.

You also need to create a Google Cloud Storage bucket. Create it, make it only accessible to your identity. Tresor will attempt to authenticate with Google by using application-default credentials.

## How to use it?

Tresor can tell you how to use it!

```
tresor help
```
