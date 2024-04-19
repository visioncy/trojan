set CGO_ENABLED=0
set GOARCH=amd64
set GOOS=linux

set version=2.15.3
set goversion=1.21.9
set gitversion=2.39.1
go build -ldflags "-w -s -X 'trojan/trojan.MVersion=%version%' -X 'trojan/trojan.BuildDate=%DATE%' -X 'trojan/trojan.GoVersion=%goversion%' -X 'trojan/trojan.GitVersion=%gitversion%'" -o "result/trojan" .
rem go build -ldflags "-w -s -X 'trojan/trojan.MVersion=%version%' -X 'trojan/trojan.BuildDate=%DATE%' -X 'trojan/trojan.GoVersion=%goversion%' -X 'trojan/trojan.GitVersion=%gitversion%'" -o "result/trojan-linux-amd64" .
