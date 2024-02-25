## Scout

Scout is a Next.js page dependency generator. This tool is aimed identify which pages are affect by file changes within your source files. Simply pass in a list of files that have changed and they will be compared against the dependcies of all page

#### Why?

QA time is valuable. Rather than having to worry about full regressions for any new work, you can identify a subset of affect pages. This will help make QA testing more effective and efficient.

### Usage

Build

```
make build
```

### Dependency Check

This task will help indentify pages that will be affect by file changes.

Using tool on uncommited changes

```
git status -s | awk '{print $2}' | xargs <path_to_binary> -p <path_to_next_app>
```

Using tool against two branches

```
git diff <branh_you want_to_check>..<base_branch> --name-only | xargs <path_to_binary> -p <path_to_next_app>
```

### Dead Code Check

This task will help indentify files not being used by any of tha pages in the app. Any `.md` or `.test.` files will be ignored

```
<path_to_binary> -p <path_to_next_app> -t deadcode-check
```
