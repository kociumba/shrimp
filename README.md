# shrimp

shrimp is a minimal cross-platform CLI profile manager for arbitrary files, particularly useful for configs.

Have you ever set up a window manager, and wanted to have multiple configs at once. This is exactly what shrimp solves.

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

To install shrimp use go:

```shell
go install github.com/kociumba/shrimp
```

<details>
    <summary>Install from source</summary>
    If you want to avoid installing stale versions from the go servers using `go install`, you can clone the repo and use `go install .` in the root of the cloned repo.
</details>
</br>

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
    > [!TIP]
    > switching is guarded by validation, so no destructive operations should run, if you want to be sure only validation is performed, use `-d` or `--dry` flags to perform a dry run
3. add files which the profile is supposed to manage:
    ```shell
    shrimp file add <path_to_file>
    ```

To see detailed info on all commands use `-h` or `--help` which displays contextual info about a command and all of it's flags

## How it works

shrimp has a very simple operating principle, it doesn't do any git integration or special storage for managed files and directories.

When a profile is deactivated all files and directories managed by it are renamed like so: `~/some.file` -> `~/some.file.profile_name.disabled`, when a profile is activated all of it's files are restored to their original names.

This simple principle makes sure the potential for data loss is as low as possible, and no complex operations have to be performed.

### Git integration ?

shrimp will never have git integration, since that requires things like managing alternate file versions, symlinks, scanning for secrets and so on.

But something like chezmoi integration might be added where shrimp can add all of it's managed configs to chezmoi automatically to allow for using it with your dotfiles.

