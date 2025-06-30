# shrimp

<img align="right" src="https://raw.githubusercontent.com/kociumba/shrimp/main/assets/shrimp.svg" alt="shrimp icon" width="150" height="150"/>

shrimp is a minimal cross-platform CLI profile manager for arbitrary files, particularly useful for configs.

Have you ever set up a window manager, and wanted to have multiple configs at once? This is exactly what shrimp solves.

> "real" user: I can't belive it's that shrimple !

shrimp works on the concept of profiles, you create them using:

```shell
shrimp create <profile_name>
```

> [!NOTE]
> The first profile you create is automatically activated.

and any files or directories managed by them are swapped out when they are activated:

```shell
shrimp activate <profile_name>
```

## Installation

### Prerequisites

You will need go installed (version 1.21.0 and up, see [Go installation guide](https://go.dev/doc/install))

### shrimp installation

Simply use `go install` to get shrimp:

```shell
go install github.com/kociumba/shrimp
```

<details>
    <summary>Install from source</summary>
    If you want to avoid installing stale versions from the go servers using <code>go install</code>, you can clone the repo and use <code>go install .</code> in the root of the cloned repo.
</details>

Prebuilt binaries may also be available in releases, depending on when you are reading this 😉

## Usage

Using shrimp is very simple

1. create a profile:
    ```shell
    shrimp create <profile_name>
    ```
2. switch to that profile:
    ```shell
    shrimp activate <profile_name>
    ```
    > switching is guarded by validation, so no destructive operations should run, if you want to be sure only validation is performed, use `-d` or `--dry` flags to perform a dry run
3. add files which the profile is supposed to manage:
    ```shell
    shrimp file add <path_to_file>
    ```

To see detailed info on all commands, use `-h` or `--help` which displays contextual info about a command and all of its flags

### Scope of the config

By default shrimp uses a global config in `~/.config/shrimp/shrimp.toml`, this has the issue that every profile needs to account for any file that might change between them, as shrimp disables all files in a profile when it's deactivated.

Let's say you want to manage a local `.editorconfig` or configs in `.vscode`, adding them to the global profiles would require you to add them to *every* profile where you need them

This issue is solved using the `-c` or `--config=PATH` flag, you can provide it with any command like so:

```shell
shrimp c <profile_name> -c ./config # uses the shorthand c for create
```

This would create a profile as usual, but it would be stored and operated from a `./config/shrimp.toml` file instead of the global one.

Think of it this way: profiles within a config are exclusive (only one profile's files can be active at a time). By using a local config you can avoid running into issues with the exclusivity at the global level

### Hooks

Some config changes need to be reloaded or can only be performed when an app is not running, to achieve this you can use pre/post activate hooks in shrimp

To set them for the active profile using the `hook` command:

```shell
shrimp hook pre -- sh -c "echo pre activation hook"
```

This method is good if you want to set up quick and simple hooks, for more complex pre and post hooks it is recommended to use shell files like so:

```shell
shrimp hook post -- pwsh -NoProfile -File ~/scripts/cleanup.ps1
```

It is recommended to use `--` to pass these commands to shrimp since this allows shrimp to freely treat everything after `--` as part of the command.

## How it works

shrimp has a very simple operating principle, it doesn't do any git integration or special storage for managed files and directories.

When a profile is deactivated all files and directories managed by it are renamed like so: `~/some.file` -> `~/some.file.profile_name.disabled`, when a profile is activated all of its files are restored to their original names.

This simple approach ensures the potential for data loss is minimal and avoids complex operations.

### Git integration ?

shrimp will never have git integration, since that requires things like managing alternate file versions, symlinks, scanning for secrets and so on.

But something like chezmoi integration might be added where shrimp can add all of its managed configs to chezmoi automatically to allow for using it with your dotfiles.

