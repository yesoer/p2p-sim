# P2PSim
 A simplified environment to test, develop and analyze peer 2 peer systems/algorithms.

 Not production ready but under active development.

## Table of Contents
- [Build](#build)
- [Run](#run)
- [Features to be Implemented](#features-to-be-implemented)
- [Contribution](#contribution)
  - [Branch Naming](#branch-naming)
  - [Commit Messages](#commit-messages)
  - [Other Tips/Notes on Contributing](#other-tipsnotes-on-contributing)

## Build

```sh
go build main.go
```

## Run

```sh
./main
```

## Features to be Implemented

Topics to implement (no specific order) : 

- Draw edges/connections using drag and drop
- Show transmitted data along edges
- Code generation for the networking part
- Define connections using a go function e.g. to connect nodes depending on the custom data/ids
- Debug option which means communication is synced (all nodes keep working until they reach either a send or await)
- Intermediate logs (e.g. via streaming, see TODO in Node.Run())
- Stress test functionality with varying configurations.
- More ways to design the network : 
  - quick connect : chord ring, tree (random or binary)
  - non-directed edges etc.

Topics to look into (whether we want them) :
- Port numbering model ? LOCAL model ?
- Automated proof generation (probably requires restriction to certain instructions etc.)

Minimizing what users have to do and optimising how they do it, is a continuous effort.

## Contribution

When confused, no worries, just publish what you have. 
A not perfectly structured contribution is still far better than nothing.

### Branch Naming 
Branch names should look like this
`<type>/<name>`
where words in "name" are separated by '-'
and type is one of the following (extend if needed)

| type | when to use      |
|------|------------------|
| feat | any new features |
| maintenance | any work on docs, git workflows, tests etc. |
| refactor | when refactoring existing parts of the application |
| fix  | bug fixes        |
| test | testing environments/throwaway branches |

### Commit Messages
More specific distinction happens in **commit messages** which should be structured
as follows :

```
<type>(<scope>): <subject>
```

- **type**
Must be one of the following:

  * **feat**: A new feature
  * **fix**: A bug fix
  * **docs**: Documentation only changes
  * **style**: Changes that do not affect the meaning of the code (white-space, formatting, missing
    semi-colons, etc)
  * **refactor**: A code change that neither fixes a bug nor adds a feature
  * **perf**: A code change that improves performance
  * **test**: Adding missing or correcting existing tests
  * **chore**: Changes to the build process or auxiliary tools and libraries such as documentation
  generation

- **scope** refers to the part of the software, which usually will be best identified by the package name.

- **subject** gives a short idea of what was done/what the intend of the commit is.

As for the **commit body** there is no mandatory structure as of now.

### Other Tips/Notes on Contributing

**Issues and Pull Requests** for now will not have any set guidelines.

As a rule of thumb for **merging** make sure to rebase before doing so.

When developing or testing you should consider using the -race flag to make sure
your contribution is fine.
Eventhough your code might seem to run flawlessly, it might be buggy.

Check out [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for commmon code review topics around golang.

