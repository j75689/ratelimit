## Local Run
```bash
make run
```

## Local Test
```bash
for (( i=0 ; i<70 ; i++ )); do; curl http://localhost:8080/ratelimit && echo ""; done
```

## Demo
```bash
for (( i=0 ; i<70 ; i++ )); do; curl https://ratelimit.retsa.tw/ratelimit && echo ""; done
```