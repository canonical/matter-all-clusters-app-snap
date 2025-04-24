# Matter All Clusters App Snap

The Matter/CHIP All Clusters App implements all the Matter clusters.
It can be used to simulate a Matter device on Linux.

## Installation

Install from the store:
```shell
sudo snap install matter-all-clusters-app
```

Connect required interfaces:
```shell
sudo snap connect matter-all-clusters-app:avahi-control
sudo snap connect matter-all-clusters-app:bluez
```

## Development

### Build the snap

Build locally for the same architecture as the host:

```bash
snapcraft -v
```

Build remotely for all supported architectures:

```bash
snapcraft remote-build
```

### Install the built snap

Install the local snap:

```bash
sudo snap install --dangerous *.snap
```

Also connect the interfaces as described under [installation](#installation).

## Test

Refer to [tests](./tests).