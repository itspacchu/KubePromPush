all: copy

copy:
	kubectl cp ./ monitoring/gotest:/root/ -c gotest
	kubectl exec -n monitoring gotest -c gotest -- sh -c "cd /root/ && go build"
	kubectl exec -n monitoring gotest -c gotest -- sh -c "cd /root/ && ./node-exporter-pusher"
