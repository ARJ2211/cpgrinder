<p align="center">
  <img src="assets/logo.png" alt="CPGrinder logo" width="700">
</p>

<p align="center">
  A terminal-first competitive programming workspace built in Go
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/License-Apache_2.0-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Problems-8200+-orange?style=flat-square&logo=codeforces&logoColor=white" />
  <img src="https://img.shields.io/badge/Platform-Terminal-black?style=flat-square&logo=gnometerminal&logoColor=white" />
</p>

## Demo

<p align="center">
  <img src="assets/process.gif" alt="CPGrinder demo">
</p>

---

## What Is CPGrinder?

CPGrinder is a TUI (terminal user interface) built in Go for daily competitive programming practice. It keeps your entire workflow in one place: browsing problems, opening starter files in your editor, running sample test cases, and reviewing your run history - all without switching windows or contexts.

The problem bank is sourced from **Codeforces**, with over **8,200 problems** scraped and available locally - so you can browse, filter, and solve without ever opening a browser.

---

## Features

| Feature                | Description                                                                           |
| ---------------------- | ------------------------------------------------------------------------------------- |
| **Problem Bank**       | 8,200+ Codeforces problems with metadata: title, source, difficulty, topics, and tags |
| **Problem Detail**     | Full statement view with rendered markdown and sample I/O                             |
| **Editor Integration** | Press `e` to open your solution file in your configured editor                        |
| **Sample Execution**   | Press `r` to run your solution against stored sample test cases                       |
| **Language Switching** | Press `l` to switch between supported languages per problem                           |
| **Run History**        | Press `a` to view all previous attempts for the current problem                       |
| **Filter & Search**    | Narrow problems by source, difficulty, title, topic, and tag                          |

### Supported Languages

CPGrinder ships starter templates and full run support for all five languages. Select your language per problem at any time with `l`.

| Language   | Runtime          | Identifier   |
| ---------- | ---------------- | ------------ |
| Python 3   | `python3`        | `python3`    |
| JavaScript | Node.js          | `javascript` |
| C++        | `g++`            | `cpp`        |
| Go         | `go run`         | `go`         |
| Java       | `javac` / `java` | `java`       |

---

## Current Status

**Working now:**

- Problem browsing, filtering, and search across 8,200+ Codeforces problems
- Full problem statement viewer with samples
- Editor integration and starter templates
- Sample test execution
- Language switching per problem
- Per-question run history

**In progress:**

- Global activity dashboard across all problems _(foundation is stable; this is the next stage)_

---

## Keybindings

| Key               | Action                               |
| ----------------- | ------------------------------------ |
| `enter` / `space` | Open selected problem                |
| `tab`             | Switch focus between list and detail |
| `f`               | Open filters                         |
| `e`               | Open solution file in editor         |
| `r`               | Run sample tests                     |
| `l`               | Open language selection              |
| `a`               | Show run history for current problem |
| `esc`             | Close overlay or go back             |
| `n` / `b`         | Next / previous page                 |
| `q`               | Quit                                 |

### Filter & Search

Press `f` from the problem list to open the filter overlay. You can narrow the problem bank by any combination of fields:

| Field          | How it works                                          |
| -------------- | ----------------------------------------------------- |
| **Title**      | Free-text search - type any part of the problem title |
| **Source**     | Cycle through available sources with `↑` / `↓`        |
| **Difficulty** | Cycle through difficulty levels with `↑` / `↓`        |
| **Topic**      | Free-text search - type a topic keyword               |
| **Tag**        | Free-text search - type a tag keyword                 |

**Filter keybindings:**

| Key         | Action                                        |
| ----------- | --------------------------------------------- |
| `tab`       | Move to next field                            |
| `shift+tab` | Move to previous field                        |
| `↑` / `↓`   | Cycle options on Source and Difficulty fields |
| `enter`     | Apply filters                                 |
| `esc`       | Close without applying                        |
| `ctrl+r`    | Reset all filters                             |
| `ctrl+u`    | Clear the current field                       |

### Language Selection

Press `l` from any problem to open the language picker. Use `↑` / `↓` to move between options and `enter` to select. The active language is marked with `[*]`. Press `esc` to cancel without changing.

Available options shown in the picker:

```
> [*] Python 3   (python3)
  [ ] JavaScript (javascript)
  [ ] C++        (cpp)
  [ ] Go         (go)
  [ ] Java       (java)
```

Your selection is saved per-problem - switching languages on one problem does not affect others.

---

## Getting Started

### Prerequisites

- Go installed
- A workspace directory configured
- An editor available in your environment

### Run

```bash
go run ./cmd/cpgrinder
```

### Build

```bash
go build -o cpgrinder ./cmd/cpgrinder
./cpgrinder
```

---

## How Solutions Work

CPGrinder judges your solution the same way competitive programming platforms do:

- Your program reads input from **stdin**
- Your program writes its answer to **stdout**
- CPGrinder captures stdout and compares it against the expected output

> **The goal is not to return a value - it is to print the answer.**

The most common mistake is writing logic that returns a value but never prints it. If nothing is written to stdout, CPGrinder sees empty output and marks it wrong.

---

### Correct vs Incorrect Patterns

#### Python

```python
# ✅ Correct
import sys
data = sys.stdin.read().strip().split()
a, b = int(data[0]), int(data[1])
print(a + b)

# ❌ Incorrect - nothing is printed
def solve():
    return 5
solve()
```

#### JavaScript

```javascript
// ✅ Correct
const fs = require("fs");
const input = fs.readFileSync(0, "utf8").trim().split(/\s+/);
console.log(parseInt(input[0]) + parseInt(input[1]));

// ❌ Incorrect - return value is never printed
function solve() {
    return 5;
}
solve();
```

#### C++

```cpp
// ✅ Correct
#include <iostream>
using namespace std;
int main() {
    int a, b;
    cin >> a >> b;
    cout << (a + b) << "\n";
}

// ❌ Incorrect - solve() returns but nothing is printed
#include <iostream>
using namespace std;
int solve() { return 5; }
int main() {
    solve();
}
```

#### Go

```go
// ✅ Correct
package main

import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    in := bufio.NewReader(os.Stdin)
    var a, b int
    fmt.Fscan(in, &a, &b)
    fmt.Println(a + b)
}

// ❌ Incorrect - result is returned but never printed
package main

func solve() int { return 5 }

func main() {
    solve()
}
```

#### Java

```java
// ✅ Correct
import java.util.*;
public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int a = sc.nextInt(), b = sc.nextInt();
        System.out.println(a + b);
    }
}

// ❌ Incorrect - return value is never printed
public class Main {
    static int solve() { return 5; }
    public static void main(String[] args) {
        solve();
    }
}
```

### Output Rules

- Trailing whitespace differences are normalized
- Wrong values, missing lines, or wrong ordering will fail
- Best practice: print exactly what the problem asks for, ending with a newline

---

## Typical Workflow

1. Select a problem from the list
2. Press `e` to open the solution file in your editor
3. Write code that reads from stdin and prints to stdout
4. Save and return to CPGrinder
5. Press `r` to run the sample tests
6. Fix issues until all samples pass
7. Press `a` to review your run history for that problem

---

## Project Structure

```
.
├── cmd
│   └── cpgrinder
│       └── main.go
├── internal
│   ├── platform
│   ├── solve
│   │   ├── templates
│   │   ├── compare.go
│   │   ├── config.go
│   │   ├── lang.go
│   │   ├── runner.go
│   │   ├── samples.go
│   │   ├── templates.go
│   │   └── workspace.go
│   ├── store
│   │   ├── fixtures
│   │   ├── store.go
│   │   ├── types.go
│   │   └── types_json.go
│   └── textlite
├── tui
│   ├── attempts
│   ├── filtersearch
│   ├── problemdetail
│   ├── problemlist
│   ├── styles
│   └── model.go
├── README.md
└── go.mod
```

---

## Architecture

| Package             | Responsibility                                                  |
| ------------------- | --------------------------------------------------------------- |
| `internal/store`    | SQLite-backed problem catalog and attempt persistence           |
| `internal/solve`    | Workspaces, templates, language selection, sample execution     |
| `tui/problemlist`   | Main browsing and filtering experience                          |
| `tui/problemdetail` | Statement view, editor launch, sample runs, per-problem actions |
| `tui/attempts`      | Run history view for the current problem                        |

CPGrinder keeps a local SQLite store for problems and attempts, and a per-problem workspace on disk for code files and configuration. When you run samples, it compiles or executes your solution and stores the outcome as an attempt record - giving you a persistent history instead of a throwaway run.

---

## Roadmap

- [ ] Global activity dashboard across all solved and attempted problems
- [ ] Richer attempt analytics
- [ ] Daily practice streaks and workflows
- [ ] Stronger progress tracking
- [ ] Improved workspace and session management

---

## License

Apache License 2.0
