

# SSH Benchmark

> A server benchmarking tool that supports offline and batch execution.

---

## Explanation

Current tool implementation process description:

1. Check if the host configuration is correct.
2. Check if the host is connectable (connectivity test).
3. Launch Iperf3 server-side.
4. Stress test script uploaded to target server and execute the script in batches according to the host list (executed serially).
5. Actual and display of results.
6. Clean up script data (complete).

---

## Tool packaging and compilation.

```bash
go mod tidy
go build -o ssh-benchmark main.go
```

If you have installed the [task](https://taskfile.dev/), you can also use it.

```bash
task build:binary
```

---

## Instructions for use

1. Modify the configuration in `config.json`

   > 1. If you don't have a license for Geekbench 5 for Linux, you can contact me by email (<yangzun@treesir.pub>). I'd be willing to share it with you.
   > 2. If you want to perform simultaneous testing on multiple nodes, fill in the `Host` field with a comma-separated list.


2. Execute the pre-compiled ssh-benchmark file of this project

   ```bash
   ./ssh-benchmark
   ```
   
    ![2023-07-31 18.15.23 2](images/README/2023-07-31%2018.15.23%202.gif)
