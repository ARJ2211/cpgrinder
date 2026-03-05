# How to Write Solutions in CPGrinder (Stdout, Not Return Values)

When you solve a problem in CPGrinder, your program is judged the same way Codeforces/LeetCode-style platforms judge it:

- Your program reads input from **stdin**
- Your program writes the final answer to **stdout**
- CPGrinder compares what you printed with the expected output (with minor whitespace normalization)

So the goal is not to return a value to CPGrinder. The goal is to print the answer.

## The Mental Model

Think of your code as a small command-line program.

Input arrives like the user typed it into the terminal.

You produce output by printing.

CPGrinder runs your program, captures what it printed, and checks it.

## What You Must Do

1. Read input

- Use stdin reading patterns in your language.

2. Compute the answer

3. Print the answer

- Exactly what the problem asks for.
- Usually one line, sometimes multiple lines.

## The Most Common Mistake

People write logic that returns a value but never prints it.

If you return something inside a function but donât print it, CPGrinder sees empty output and marks it wrong.

## Language Examples

### Python

Correct

```py
import sys

data = sys.stdin.read().strip().split()
a = int(data[0]); b = int(data[1])
print(a + b)
```

Incorrect (does not print)

```py
def solve():
    return 5
solve()
```

If you use a solve() function, you still must print its result:

```py
def solve(data: str) -> str:
    return "5"

import sys
out = solve(sys.stdin.read())
print(out)
```

### JavaScript (Node)

Correct

```js
const fs = require("fs");
const input = fs.readFileSync(0, "utf8").trim().split(/\s+/);
const a = parseInt(input[0], 10);
const b = parseInt(input[1], 10);
console.log(a + b);
```

Incorrect (does not print)

```js
function solve() {
    return 5;
}
solve();
```

If your template returns a value, print it:

```js
function solve() {
    return 5;
}
const out = solve();
process.stdout.write(String(out) + "\n");
```

### C++

Correct

```cpp
#include <iostream>
using namespace std;

int main() {
    int a, b;
    cin >> a >> b;
    cout << (a + b) << "\n";
}
```

Incorrect (computes but never prints)

```cpp
int solve() { return 5; }
int main() { solve(); }
```

### Go

Correct

```go
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
```

### Java

Correct

```java
import java.util.*;

public class Main {
  public static void main(String[] args) {
    Scanner sc = new Scanner(System.in);
    int a = sc.nextInt();
    int b = sc.nextInt();
    System.out.println(a + b);
  }
}
```

## Output Rules CPGrinder Uses

- Your output is compared against expected output.
- Extra trailing spaces and extra blank lines usually donât matter (CPGrinder normalizes).
- But missing lines, wrong values, or wrong order will fail.

Best practice:

- Print exactly what the problem asks for.
- End your output with a newline.

## Workflow in CPGrinder

1. Select a problem
2. Press `e` to open the solution file in your editor
3. Write code that reads stdin and prints to stdout
4. Save and close the editor
5. Press `r` to run samples
6. Fix until all samples pass
