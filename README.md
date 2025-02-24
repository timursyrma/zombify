# Zombie Process Daemon

⚠️ Warning. Running this tool may impact system stability and resource availability. Use with caution and only in controlled environments.

## Build

```bash
go build -o zombie_daemon cmd/zombie_daemon.go
```

## Usage

### Start daeёmon
```bash
sudo ./zombie_daemon
```

### Monitor zombies
```bash
# Check zombie processes
ps aux | grep -w Z

# Monitor logs
tail -f /tmp/zombie_daemon.log
```

### Stop daemon
```bash
sudo kill $(cat /tmp/zombie_daemon.pid)
```

### If you get RLIMIT error
```bash
sudo sh -c 'ulimit -u 2500'
``` 