# Pr!nt: The Cloud-Netboot ISO Builder

Pr!nt is a tool for building small, bootable ISO images that install operating systems over the network.

Instead of bundling complete operating system images into a large ISO, Pr!nt creates a lightweight boot environment that connects to the network and retrieves the required installation resources when needed.

Pr!nt is designed to provide a single bootable environment for installing supported Linux and BSD distributions without maintaining a collection of large, pre-downloaded ISO images.

## Features

* Build lightweight bootable ISOs for network-based operating system installation
* Install supported Linux and BSD distributions directly over the network
* Interactive terminal user interface (TUI) for selecting an operating system and configuring installation options
* Select installation mirrors based on geographic location
* Configure Wi-Fi before booting the generated ISO
* Support for Ethernet and Wi-Fi network connections
* Build ISOs entirely from the command line
* List supported operating systems from the CLI
* Write generated ISOs directly to USB devices
* Require explicit confirmation before performing destructive USB writes
* Use distribution-specific installation environments where available
* Written in Go
* Open source and licensed under the Apache License 2.0

## Supported Operating Systems

| Operating System | Installation Environment |
| ---------------- | ------------------------ |
| Debian           | `debian-installer`       |
| Ubuntu           | `debian-installer`       |
| Fedora           | Anaconda / Netboot       |
| Arch Linux       | `archinstall`            |
| Alpine Linux     | Alpine Setup             |
| FreeBSD          | `bsdinstall`             |
| OpenBSD          | `bsd.rd`                 |
| openSUSE         | Linux + Initrd           |
| NixOS            | NixOS Installer          |
| Rocky Linux      | Anaconda                 |
| AlmaLinux        | Anaconda                 |
| Oracle Linux     | Dracut                   |
| Void Linux       | Live ISO                 |
| Gentoo           | LiveCD ISO               |
| Clear Linux      | Netboot Tarball          |

Windows and macOS are not supported due to licensing and distribution restrictions.

## Usage

### Interactive mode

Launch the interactive TUI:

```sh
go run ./cmd/print
```

The interactive workflow allows you to select the operating system, country, mirror, and network configuration before generating the ISO.

### List supported operating systems

```sh
go run ./cmd/print -list-distros
```

### Build a Debian ISO

Build a Debian ISO with preconfigured Wi-Fi:

```sh
go run ./cmd/print \
    -distro debian \
    -country DE \
    -wifi-ssid "MyNet" \
    -wifi-pass "secret" \
    -out debian.iso
```

### Build and write an ISO to USB

```sh
go run ./cmd/print \
    -distro fedora \
    -out fedora.iso \
    -write \
    -device /dev/sdb \
    -yes
```

> **Warning:** Writing directly to a block device can permanently destroy data on the target device. Verify the target device before using the `-write` option.

## Building

Pr!nt requires Go 1.24 or newer.

ISO generation requires either `grub-mkrescue` or `xorriso` to be available on your `PATH`.

### Build the project

```sh
go build ./...
```

### Run the test suite

```sh
go test ./...
```

### Run Pr!nt from source

```sh
go run ./cmd/print
```

## How It Works

A typical Pr!nt installation follows this process:

1. **Build** - Generate a small bootable ISO.
2. **Flash** - Write the ISO to a USB device.
3. **Boot** - Start the target computer from the USB device.
4. **Connect** - Establish a network connection using Ethernet or Wi-Fi.
5. **Select** - Choose an operating system using the TUI.
6. **Install** - Retrieve the required installation resources and start the operating system's installer.

Pr!nt does not bundle complete operating system images into the generated ISO. The required installation resources are retrieved over the network during the installation process.

## Troubleshooting

If you encounter a problem with Pr!nt, please open an issue in the project's GitHub repository.

When reporting a problem, include:

* The Pr!nt version or Git commit
* Operating system and architecture
* Go version
* The command used
* Relevant terminal output or logs
* Steps required to reproduce the issue

Do not include sensitive information such as Wi-Fi passwords, API keys, or other credentials in issue reports.

## Development

Contributions are welcome.

Before submitting a pull request, run:

```sh
go build ./...
go test ./...
```

Please keep changes focused and include tests for new functionality where appropriate.

For larger changes, consider opening an issue before beginning implementation.

## Security

Pr!nt performs operations that can interact directly with storage devices.

If you discover a security vulnerability, please report it privately through the project's security reporting process rather than publicly disclosing the issue.

Never include credentials, private keys, Wi-Fi passwords, or other sensitive information in bug reports or logs shared publicly.

## Additional Information

Pr!nt is maintained by the **Pr!nt Foundation**.

* Source code: GitHub repository
* Issue tracker: GitHub Issues
* License: Apache License 2.0

## License

Pr!nt is licensed under the [Apache License 2.0](LICENSE).

Copyright © 2026 Pr!nt Foundation.
